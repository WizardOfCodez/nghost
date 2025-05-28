package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"nghost/internal/config"
	"nghost/internal/nkn"
	"nghost/internal/vpn"
)

func main() {
	var (
		configPath = flag.String("config", "config.json", "Path to configuration file")
		exitNode   = flag.Bool("exit-node", false, "Run as exit node")
		daemon     = flag.Bool("daemon", false, "Run as daemon (headless)")
		listPeers   = flag.Bool("list-peers", false, "List discovered peers")
		addPeer     = flag.String("add-peer", "", "Add peer by NKN address")
		exportPeers = flag.String("export-peers", "", "Export peers to JSON file")
		status      = flag.Bool("status", false, "Show network interface status")
		testDiscovery = flag.Bool("test-discovery", false, "Test exit node discovery")
	)
	flag.Parse()

	// Handle peer management commands
	if *listPeers {
		if err := listPeersCmd(*configPath); err != nil {
			log.Fatalf("Failed to list peers: %v", err)
		}
		return
	}

	if *addPeer != "" {
		if err := addPeerCmd(*configPath, *addPeer); err != nil {
			log.Fatalf("Failed to add peer: %v", err)
		}
		return
	}

	if *exportPeers != "" {
		if err := exportPeersCmd(*configPath, *exportPeers); err != nil {
			log.Fatalf("Failed to export peers: %v", err)
		}
		return
	}

	if *status {
		showNetworkStatus()
		return
	}

	if *testDiscovery {
		if err := testExitNodeDiscovery(*configPath); err != nil {
			log.Fatalf("Failed to test discovery: %v", err)
		}
		return
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	nknClient, err := nkn.NewClient(cfg.NKN)
	if err != nil {
		log.Fatalf("Failed to create NKN client: %v", err)
	}
	defer nknClient.Close()

	vpnEngine, err := vpn.NewEngine(cfg.VPN, nknClient)
	if err != nil {
		log.Fatalf("Failed to create VPN engine: %v", err)
	}

	if *exitNode {
		fmt.Println("Starting NGhost as exit node...")
		if err := vpnEngine.StartExitNode(); err != nil {
			log.Fatalf("Failed to start exit node: %v", err)
		}
	} else if *daemon {
		fmt.Println("Starting NGhost daemon...")
		if err := vpnEngine.StartDaemon(); err != nil {
			log.Fatalf("Failed to start daemon: %v", err)
		}
	} else {
		fmt.Println("Starting NGhost GUI...")
		// TODO: Launch GUI application
		fmt.Println("GUI not implemented yet, running in daemon mode")
		if err := vpnEngine.StartDaemon(); err != nil {
			log.Fatalf("Failed to start daemon: %v", err)
		}
	}

	select {} // Keep running
}

// Helper functions for peer management commands
func listPeersCmd(configPath string) error {
	return listPeers(configPath)
}

func addPeerCmd(configPath, peerAddr string) error {
	return addPeer(configPath, peerAddr)
}

func exportPeersCmd(configPath, outputPath string) error {
	return exportPeers(configPath, outputPath)
}

func listPeers(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	nknClient, err := nkn.NewClient(cfg.NKN)
	if err != nil {
		return fmt.Errorf("failed to create NKN client: %w", err)
	}
	defer func() {
		// Ensure clean shutdown
		if cerr := nknClient.Close(); cerr != nil {
			fmt.Printf("Warning: error closing NKN client: %v\n", cerr)
		}
	}()

	fmt.Println("NGhost Peer Status")
	fmt.Printf("Your NKN Address: %s\n\n", nknClient.GetAddress())

	// Give client a moment to connect
	time.Sleep(2 * time.Second)

	peers := nknClient.GetPeers()
	if len(peers) == 0 {
		fmt.Println("No peers discovered yet.")
		fmt.Println("\n💡 To add peers manually:")
		fmt.Println("   ./nghost -add-peer <nkn-address>")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ADDRESS\tIP\tSTATUS\tEXIT NODE\tLAST SEEN")
	fmt.Fprintln(w, "-------\t--\t------\t---------\t---------")

	for _, peer := range peers {
		status := "offline"
		if peer.Online {
			status = "online"
		}
		exitNode := "no"
		if peer.ExitNode {
			exitNode = "yes"
		}
		lastSeen := peer.LastSeen.Format("15:04:05")
		if peer.LastSeen.IsZero() {
			lastSeen = "never"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			peer.Address[:16]+"...",
			peer.IPAddress,
			status,
			exitNode,
			lastSeen)
	}
	w.Flush()

	exitNodes := nknClient.FindExitNodes()
	if len(exitNodes) > 0 {
		fmt.Printf("\nAvailable exit nodes: %d\n", len(exitNodes))
	}

	return nil
}

func addPeer(configPath, peerAddr string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	nknClient, err := nkn.NewClient(cfg.NKN)
	if err != nil {
		return fmt.Errorf("failed to create NKN client: %w", err)
	}
	defer nknClient.Close()

	nknClient.AddPeer(peerAddr)
	fmt.Printf("Added peer: %s\n", peerAddr)

	return nil
}

func exportPeers(configPath, outputPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	nknClient, err := nkn.NewClient(cfg.NKN)
	if err != nil {
		return fmt.Errorf("failed to create NKN client: %w", err)
	}
	defer nknClient.Close()

	peers := nknClient.GetPeers()
	data, err := json.MarshalIndent(peers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal peers: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Exported %d peers to %s\n", len(peers), outputPath)
	return nil
}