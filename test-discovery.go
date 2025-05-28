package main

import (
	"fmt"
	"time"

	"nghost/internal/config"
	"nghost/internal/nkn"
)

func testExitNodeDiscovery(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("🔍 Testing exit node discovery...")
	
	nknClient, err := nkn.NewClient(cfg.NKN)
	if err != nil {
		return fmt.Errorf("failed to create NKN client: %w", err)
	}
	defer nknClient.Close()

	fmt.Printf("📡 Your NKN Address: %s\n", nknClient.GetAddress())
	fmt.Println("🔗 Waiting for NKN connection...")
	
	// Wait for connection and discovery
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		
		peers := nknClient.GetPeers()
		exitNodes := nknClient.FindExitNodes()
		
		fmt.Printf("\r⏱️  %02ds - Peers: %d, Exit nodes: %d", i+1, len(peers), len(exitNodes))
		
		if len(exitNodes) > 0 {
			fmt.Printf("\n🎉 Found exit nodes!\n")
			for _, node := range exitNodes {
				fmt.Printf("  🚪 %s (IP: %s)\n", node.Address[:16]+"...", node.IPAddress)
			}
			return nil
		}
	}
	
	fmt.Printf("\n❌ No exit nodes discovered after 30 seconds\n")
	fmt.Println("💡 Make sure an exit node is running with: sudo ./nghost -exit-node")
	return nil
}