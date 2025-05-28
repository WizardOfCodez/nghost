//go:build linux || darwin

package tun

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	IFF_TUN   = 0x0001
	IFF_NO_PI = 0x1000
	TUNSETIFF = 0x400454ca
)

func (d *Device) createLinux() error {
	fd, err := syscall.Open("/dev/net/tun", syscall.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open /dev/net/tun: %w", err)
	}

	var ifr struct {
		name  [16]byte
		flags uint16
		_     [22]byte
	}

	copy(ifr.name[:], d.name)
	ifr.flags = IFF_TUN | IFF_NO_PI

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), TUNSETIFF, uintptr(unsafe.Pointer(&ifr)))
	if errno != 0 {
		syscall.Close(fd)
		return fmt.Errorf("failed to create TUN device: %v", errno)
	}

	d.fd = fd
	return nil
}


func (d *Device) configureLinux(ip, network string) error {
	commands := [][]string{
		{"ip", "addr", "add", ip + "/16", "dev", d.name},
		{"ip", "link", "set", "dev", d.name, "up"},
		{"ip", "link", "set", "dev", d.name, "mtu", fmt.Sprintf("%d", d.mtu)},
	}

	for _, cmd := range commands {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			return fmt.Errorf("failed to run command %v: %w", cmd, err)
		}
	}
	return nil
}


func (d *Device) readUnix() ([]byte, error) {
	buf := make([]byte, 65536)
	n, err := syscall.Read(d.fd, buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (d *Device) writeUnix(packet []byte) error {
	_, err := syscall.Write(d.fd, packet)
	return err
}

func (d *Device) closeUnix() error {
	if d.fd > 0 {
		return syscall.Close(d.fd)
	}
	return nil
}