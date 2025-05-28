package tun

import (
	"fmt"
	"net"
	"runtime"
)

type Device struct {
	name           string
	cidr           string
	mtu            int
	fd             int
	simulationMode *SimulationDevice
}

func NewDevice(name, cidr string, mtu int) (*Device, error) {
	device := &Device{
		name: name,
		cidr: cidr,
		mtu:  mtu,
	}

	if err := device.create(); err != nil {
		// If TUN creation fails, offer simulation mode
		fmt.Printf("❌ Failed to create real TUN device: %v\n", err)
		fmt.Printf("🔄 Falling back to simulation mode for testing...\n")
		
		simDevice, simErr := NewSimulationDevice(name, cidr, mtu)
		if simErr != nil {
			return nil, fmt.Errorf("both real and simulation TUN failed: real=%v, sim=%v", err, simErr)
		}
		
		// Return a device that wraps the simulation
		return &Device{
			name:           name,
			cidr:           cidr,
			mtu:            mtu,
			fd:             -1, // Mark as simulation
			simulationMode: simDevice,
		}, nil
	}

	if err := device.configure(); err != nil {
		device.Close()
		return nil, err
	}

	return device, nil
}

func (d *Device) create() error {
	switch runtime.GOOS {
	case "linux":
		return d.createLinux()
	case "darwin":
		return d.createDarwin()
	case "windows":
		return fmt.Errorf("Windows support not yet implemented - please use WSL")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func (d *Device) configure() error {
	_, network, err := net.ParseCIDR(d.cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %w", err)
	}

	// Get the first usable IP in the network for this interface
	ip := network.IP
	ip[len(ip)-1] = 1 // Use .1 as the interface IP

	switch runtime.GOOS {
	case "linux":
		return d.configureLinux(ip.String(), network.String())
	case "darwin":
		return d.configureDarwin(ip.String(), network.String())
	case "windows":
		return fmt.Errorf("Windows support not yet implemented - please use WSL")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func (d *Device) Read() ([]byte, error) {
	if d.simulationMode != nil {
		return d.simulationMode.Read()
	}
	
	switch runtime.GOOS {
	case "linux", "darwin":
		return d.readUnix()
	case "windows":
		return nil, fmt.Errorf("Windows support not yet implemented - please use WSL")
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func (d *Device) Write(packet []byte) error {
	if d.simulationMode != nil {
		return d.simulationMode.Write(packet)
	}
	
	switch runtime.GOOS {
	case "linux", "darwin":
		return d.writeUnix(packet)
	case "windows":
		return fmt.Errorf("Windows support not yet implemented - please use WSL")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func (d *Device) GetName() string {
	return d.name
}

func (d *Device) Close() error {
	if d.simulationMode != nil {
		return d.simulationMode.Close()
	}
	
	switch runtime.GOOS {
	case "linux", "darwin":
		return d.closeUnix()
	case "windows":
		return fmt.Errorf("Windows support not yet implemented - please use WSL")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}