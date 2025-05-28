//go:build linux

package tun

import (
	"fmt"
	"os/exec"
)

func (d *Device) createDarwin() error {
	return fmt.Errorf("Darwin TUN creation not available on Linux")
}

func (d *Device) createDarwinUtun() error {
	return fmt.Errorf("Darwin utun not available on Linux")
}

func (d *Device) configureDarwin(ip, network string) error {
	return fmt.Errorf("Darwin configuration not available on Linux")
}