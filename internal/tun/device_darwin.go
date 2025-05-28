//go:build darwin

package tun

import (
	"fmt"
	"os/exec"
	"syscall"
)

func (d *Device) createDarwin() error {
	fmt.Printf("🔧 Attempting to create TUN device on macOS...\n")
	
	// Try to use the utun implementation first
	fmt.Printf("🔧 Trying utun control interface...\n")
	if err := d.createDarwinUtun(); err == nil {
		fmt.Printf("✅ Created utun device: %s\n", d.name)
		return nil
	} else {
		fmt.Printf("❌ utun control failed: %v\n", err)
	}
	
	// On macOS, we can try to create TUN using ifconfig
	// This requires the system to have TUN/TAP support
	fmt.Printf("🔧 Trying ifconfig method...\n")
	for i := 0; i < 10; i++ {
		interfaceName := fmt.Sprintf("utun%d", i)
		
		// Try to create the interface using ifconfig
		cmd := exec.Command("ifconfig", interfaceName, "create")
		fmt.Printf("🔧 Trying to create interface: %s\n", interfaceName)
		if err := cmd.Run(); err == nil {
			// Interface created successfully, now try to open it
			tunPath := fmt.Sprintf("/dev/%s", interfaceName)
			fd, err := syscall.Open(tunPath, syscall.O_RDWR, 0)
			if err == nil {
				d.fd = fd
				d.name = interfaceName
				fmt.Printf("✅ Created TUN device: %s (fd=%d)\n", interfaceName, fd)
				return nil
			} else {
				fmt.Printf("❌ Failed to open %s: %v\n", tunPath, err)
			}
		} else {
			fmt.Printf("❌ Failed to create %s: %v\n", interfaceName, err)
		}
	}
	
	return fmt.Errorf("failed to create TUN device on macOS - this system may not support TUN interfaces. Try installing a TUN/TAP driver or use Docker/VM")
}

func (d *Device) configureDarwin(ip, network string) error {
	fmt.Printf("🔧 Configuring interface %s with IP %s\n", d.name, ip)
	
	commands := [][]string{
		{"ifconfig", d.name, ip, ip, "up"},
		{"route", "add", "-net", network, "-interface", d.name},
	}

	for _, cmd := range commands {
		fmt.Printf("🔧 Running: %v\n", cmd)
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			fmt.Printf("❌ Command failed: %v\n", err)
			// Don't fail on route command errors (might already exist)
			if cmd[0] == "route" {
				fmt.Printf("⚠️  Route command failed (may already exist), continuing...\n")
				continue
			}
			return fmt.Errorf("failed to run command %v: %w", cmd, err)
		}
		fmt.Printf("✅ Command succeeded\n")
	}
	return nil
}