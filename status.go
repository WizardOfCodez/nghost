package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func showNetworkStatus() {
	fmt.Println("🌐 Network Interface Status:")
	fmt.Println(strings.Repeat("=", 40))
	
	// Show all interfaces
	cmd := exec.Command("ifconfig", "-a")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("❌ Failed to get interface list: %v\n", err)
		return
	}
	
	interfaces := strings.Split(string(output), "\n")
	for _, line := range interfaces {
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "\t") {
			interfaceName := strings.Split(line, ":")[0]
			if strings.Contains(interfaceName, "tun") || 
			   strings.Contains(interfaceName, "utun") || 
			   strings.Contains(interfaceName, "nghost") {
				fmt.Printf("🔍 Found TUN-like interface: %s\n", interfaceName)
			}
		}
	}
	
	// Check for utun interfaces specifically
	fmt.Println("\n🔍 Checking for utun interfaces:")
	for i := 0; i < 10; i++ {
		interfaceName := fmt.Sprintf("utun%d", i)
		cmd := exec.Command("ifconfig", interfaceName)
		if err := cmd.Run(); err == nil {
			fmt.Printf("✅ Found existing interface: %s\n", interfaceName)
		}
	}
	
	// Check for nghost interface
	fmt.Println("\n🔍 Checking for nghost interface:")
	cmd = exec.Command("ifconfig", "nghost0")
	if err := cmd.Run(); err == nil {
		fmt.Printf("✅ nghost0 interface exists\n")
	} else {
		fmt.Printf("❌ nghost0 interface not found\n")
	}
	
	// Show routing table
	fmt.Println("\n🗺️  Routing table:")
	cmd = exec.Command("netstat", "-rn")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "10.100") || strings.Contains(line, "tun") || strings.Contains(line, "utun") {
				fmt.Printf("🗺️  %s\n", line)
			}
		}
	}
}