package vpn

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
)

func (e *Engine) setupDefaultRoute() error {
	switch runtime.GOOS {
	case "linux":
		return e.setupDefaultRouteLinux()
	case "darwin":
		return e.setupDefaultRouteDarwin()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func (e *Engine) setupDefaultRouteLinux() error {
	// Add route for VPN traffic through our interface
	_, network, _ := net.ParseCIDR(e.config.CIDR)
	cmd := exec.Command("ip", "route", "add", network.String(), "dev", e.config.InterfaceName)
	return cmd.Run()
}

func (e *Engine) setupDefaultRouteDarwin() error {
	// Add route for VPN traffic through our interface
	_, network, _ := net.ParseCIDR(e.config.CIDR)
	cmd := exec.Command("route", "add", "-net", network.String(), "-interface", e.config.InterfaceName)
	return cmd.Run()
}

func (e *Engine) getDefaultGateway() (string, error) {
	switch runtime.GOOS {
	case "linux":
		return e.getDefaultGatewayLinux()
	case "darwin":
		return e.getDefaultGatewayDarwin()
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func (e *Engine) getDefaultGatewayLinux() (string, error) {
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "default via") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return parts[2], nil
			}
		}
	}
	return "", fmt.Errorf("no default gateway found")
}

func (e *Engine) getDefaultGatewayDarwin() (string, error) {
	cmd := exec.Command("route", "-n", "get", "default")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "gateway:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}
	return "", fmt.Errorf("no default gateway found")
}

func (e *Engine) addPeerRoute(peerIP, nknAddr string) error {
	e.routesMu.Lock()
	defer e.routesMu.Unlock()

	// Create /32 route for the specific peer IP
	cidr := peerIP + "/32"
	e.routes[cidr] = nknAddr

	fmt.Printf("Added route: %s -> %s\n", cidr, nknAddr[:16]+"...")
	return nil
}

func (e *Engine) removePeerRoute(peerIP string) error {
	e.routesMu.Lock()
	defer e.routesMu.Unlock()

	cidr := peerIP + "/32"
	delete(e.routes, cidr)

	fmt.Printf("Removed route: %s\n", cidr)
	return nil
}