package tun

import (
	"fmt"
	"net"
	"time"
)

// SimulationDevice provides a mock TUN device for testing
type SimulationDevice struct {
	name     string
	cidr     string
	mtu      int
	packets  chan []byte
	running  bool
}

func NewSimulationDevice(name, cidr string, mtu int) (*SimulationDevice, error) {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %w", err)
	}

	sim := &SimulationDevice{
		name:    name,
		cidr:    cidr,
		mtu:     mtu,
		packets: make(chan []byte, 100),
		running: true,
	}

	fmt.Printf("🔧 Created simulation TUN device: %s (%s)\n", name, cidr)
	fmt.Printf("📝 Note: This is a simulation mode - no actual network traffic will be routed\n")

	// Start simulation packet generator
	go sim.generateSimulationTraffic()

	return sim, nil
}

func (s *SimulationDevice) generateSimulationTraffic() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for s.running {
		select {
		case <-ticker.C:
			// Generate a fake ping packet
			packet := s.createSimulationPacket()
			select {
			case s.packets <- packet:
				fmt.Printf("📡 Simulation: Generated test packet (%d bytes)\n", len(packet))
			default:
				// Channel full, skip
			}
		}
	}
}

func (s *SimulationDevice) createSimulationPacket() []byte {
	// Create a minimal IP packet structure for simulation
	packet := make([]byte, 60)
	
	// IP header (simplified)
	packet[0] = 0x45        // Version (4) + Header Length (5*4=20 bytes)
	packet[1] = 0x00        // Type of Service
	packet[2] = 0x00        // Total Length (high byte)
	packet[3] = 0x3c        // Total Length (low byte) = 60
	packet[9] = 0x01        // Protocol (ICMP)
	
	// Source IP: 10.100.0.1
	packet[12] = 10
	packet[13] = 100
	packet[14] = 0
	packet[15] = 1
	
	// Destination IP: 10.100.0.2
	packet[16] = 10
	packet[17] = 100
	packet[18] = 0
	packet[19] = 2
	
	return packet
}

func (s *SimulationDevice) Read() ([]byte, error) {
	if !s.running {
		return nil, fmt.Errorf("simulation device closed")
	}
	
	select {
	case packet := <-s.packets:
		return packet, nil
	case <-time.After(30 * time.Second):
		// Return timeout to prevent blocking
		return nil, fmt.Errorf("simulation timeout")
	}
}

func (s *SimulationDevice) Write(packet []byte) error {
	if !s.running {
		return fmt.Errorf("simulation device closed")
	}
	
	if len(packet) > 20 {
		// Parse destination IP for logging
		destIP := net.IP(packet[16:20])
		fmt.Printf("📤 Simulation: Would send packet to %s (%d bytes)\n", destIP, len(packet))
	}
	
	return nil
}

func (s *SimulationDevice) Close() error {
	s.running = false
	close(s.packets)
	fmt.Printf("🔧 Closed simulation TUN device: %s\n", s.name)
	return nil
}