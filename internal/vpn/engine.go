package vpn

import (
	"fmt"
	"net"
	"os/exec"
	"sync"
	"time"

	"nghost/internal/config"
	"nghost/internal/nkn"
	"nghost/internal/tun"
)

type Engine struct {
	config     *config.VPNConfig
	nknClient  *nkn.Client
	tunDevice  *tun.Device
	routes     map[string]string // CIDR -> NKN address mapping
	routesMu   sync.RWMutex
	running    bool
	runningMu  sync.RWMutex
	isExitNode bool
	myIP       net.IP
}

func NewEngine(cfg config.VPNConfig, nknClient *nkn.Client) (*Engine, error) {
	return &Engine{
		config:    &cfg,
		nknClient: nknClient,
		routes:    make(map[string]string),
	}, nil
}

func (e *Engine) StartDaemon() error {
	e.runningMu.Lock()
	defer e.runningMu.Unlock()

	if e.running {
		return fmt.Errorf("VPN engine already running")
	}

	// Create TUN interface
	tunDevice, err := tun.NewDevice(e.config.InterfaceName, e.config.CIDR, e.config.MTU)
	if err != nil {
		return fmt.Errorf("failed to create TUN device: %w", err)
	}
	e.tunDevice = tunDevice

	// Set our IP address
	_, network, _ := net.ParseCIDR(e.config.CIDR)
	e.myIP = network.IP
	e.myIP[len(e.myIP)-1] = 1 // Use .1 as our IP

	// Link NKN client with VPN engine
	e.nknClient.SetVPNEngine(e)

	// Start packet processing
	go e.processPackets()

	// Start peer announcements
	go e.announcePeer()

	e.running = true
	fmt.Printf("NGhost VPN started on interface %s (%s)\n", e.tunDevice.GetName(), e.config.CIDR)
	fmt.Printf("NKN address: %s\n", e.nknClient.GetAddress())
	fmt.Printf("VPN IP: %s\n", e.myIP.String())

	return nil
}

func (e *Engine) StartExitNode() error {
	e.isExitNode = true

	if err := e.StartDaemon(); err != nil {
		return err
	}

	// Enable IP forwarding
	if err := e.enableIPForwarding(); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}

	// Set up NAT for internet access
	if err := e.setupNAT(); err != nil {
		return fmt.Errorf("failed to setup NAT: %w", err)
	}

	// Set up default routing
	if err := e.setupDefaultRoute(); err != nil {
		fmt.Printf("Warning: failed to setup default route: %v\n", err)
	}

	fmt.Println("Configured as exit node - forwarding traffic to internet")
	return nil
}

func (e *Engine) processPackets() {
	for {
		packet, err := e.tunDevice.Read()
		if err != nil {
			continue
		}

		// Parse destination IP from packet
		if len(packet) < 20 {
			continue
		}

		destIP := net.IP(packet[16:20])
		
		// Check if destination is in our VPN network
		_, vpnNetwork, _ := net.ParseCIDR(e.config.CIDR)
		if vpnNetwork.Contains(destIP) {
			// Find peer route
			destAddr := e.findRoute(destIP)
			if destAddr != "" {
				// Forward to NKN peer
				if err := e.nknClient.SendPacket(destAddr, packet); err != nil {
					fmt.Printf("Failed to send packet via NKN: %v\n", err)
				}
			}
		} else if e.isExitNode {
			// Forward to internet (handled by system routing)
			continue
		} else {
			// Find exit node for internet traffic
			exitNodes := e.nknClient.FindExitNodes()
			if len(exitNodes) > 0 {
				// Use first available exit node
				if err := e.nknClient.SendPacket(exitNodes[0].Address, packet); err != nil {
					fmt.Printf("Failed to send packet via exit node: %v\n", err)
				}
			}
		}
	}
}

func (e *Engine) findRoute(ip net.IP) string {
	e.routesMu.RLock()
	defer e.routesMu.RUnlock()

	for cidr, addr := range e.routes {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return addr
		}
	}
	return ""
}

func (e *Engine) AddRoute(cidr, nknAddr string) error {
	e.routesMu.Lock()
	defer e.routesMu.Unlock()
	e.routes[cidr] = nknAddr
	return nil
}

func (e *Engine) InjectPacket(packet []byte) error {
	if e.tunDevice == nil {
		return fmt.Errorf("TUN device not initialized")
	}
	return e.tunDevice.Write(packet)
}

func (e *Engine) announcePeer() {
	// Wait for interface to be fully configured
	time.Sleep(2 * time.Second)
	
	// Initial announcement
	if err := e.nknClient.AnnouncePeer(e.myIP.String(), e.isExitNode); err != nil {
		fmt.Printf("❌ Failed initial peer announcement: %v\n", err)
	} else {
		if e.isExitNode {
			fmt.Printf("📢 Announced as exit node with IP %s\n", e.myIP.String())
		} else {
			fmt.Printf("📢 Announced peer with IP %s\n", e.myIP.String())
		}
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := e.nknClient.AnnouncePeer(e.myIP.String(), e.isExitNode); err != nil {
				fmt.Printf("❌ Failed to announce peer: %v\n", err)
			} else {
				fmt.Printf("📢 Peer announcement sent (exit_node=%v)\n", e.isExitNode)
			}
		}
	}
}

func (e *Engine) enableIPForwarding() error {
	return exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run()
}

func (e *Engine) setupNAT() error {
	// Get the actual interface name
	interfaceName := e.config.InterfaceName
	if e.tunDevice != nil {
		interfaceName = e.tunDevice.GetName()
	}

	// Set up iptables rules for NAT
	commands := [][]string{
		{"iptables", "-t", "nat", "-A", "POSTROUTING", "-s", e.config.CIDR, "-j", "MASQUERADE"},
		{"iptables", "-A", "FORWARD", "-i", interfaceName, "-j", "ACCEPT"},
		{"iptables", "-A", "FORWARD", "-o", interfaceName, "-j", "ACCEPT"},
	}

	fmt.Printf("🔧 Setting up NAT for exit node...\n")
	for _, cmd := range commands {
		fmt.Printf("🔧 Running: %v\n", cmd)
		execCmd := exec.Command(cmd[0], cmd[1:]...)
		output, err := execCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("⚠️  iptables command failed: %v\n", err)
			fmt.Printf("⚠️  Output: %s\n", string(output))
			// Continue with other rules even if one fails
		} else {
			fmt.Printf("✅ iptables rule added\n")
		}
	}
	
	fmt.Printf("✅ NAT configuration completed\n")
	return nil
}

func (e *Engine) AddPeerRoute(peerIP, nknAddr string) error {
	return e.addPeerRoute(peerIP, nknAddr)
}

func (e *Engine) RemovePeerRoute(peerIP string) error {
	return e.removePeerRoute(peerIP)
}

func (e *Engine) Stop() error {
	e.runningMu.Lock()
	defer e.runningMu.Unlock()

	if !e.running {
		return nil
	}

	if e.tunDevice != nil {
		e.tunDevice.Close()
	}

	e.running = false
	return nil
}