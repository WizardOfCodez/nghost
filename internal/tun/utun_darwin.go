//go:build darwin

package tun

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	CTLIOCGINFO = 0xc0644e03
	UTUN_CONTROL_NAME = "com.apple.net.utun_control"
)

type sockaddrCtl struct {
	scLen      uint8
	scFamily   uint8
	ssSysaddr  uint16
	scID       uint32
	scUnit     uint32
	scReserved [5]uint32
}

func (d *Device) createDarwinUtun() error {
	// Create socket for utun control
	fd, err := syscall.Socket(syscall.AF_SYSTEM, syscall.SOCK_DGRAM, 2) // SYSPROTO_CONTROL
	if err != nil {
		return fmt.Errorf("failed to create control socket: %w", err)
	}

	// Get control info
	var ctlInfo struct {
		ctlID   uint32
		ctlName [96]byte
	}
	copy(ctlInfo.ctlName[:], UTUN_CONTROL_NAME)

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), CTLIOCGINFO, uintptr(unsafe.Pointer(&ctlInfo)))
	if errno != 0 {
		syscall.Close(fd)
		return fmt.Errorf("failed to get control info: %v", errno)
	}

	// Connect to utun control
	addr := sockaddrCtl{
		scLen:     uint8(unsafe.Sizeof(sockaddrCtl{})),
		scFamily:  syscall.AF_SYSTEM,
		ssSysaddr: 2, // AF_SYS_CONTROL
		scID:      ctlInfo.ctlID,
		scUnit:    0, // Use any available unit
	}

	_, _, errno = syscall.Syscall(syscall.SYS_CONNECT, uintptr(fd), uintptr(unsafe.Pointer(&addr)), uintptr(unsafe.Sizeof(addr)))
	if errno != 0 {
		syscall.Close(fd)
		return fmt.Errorf("failed to connect to utun control: %v", errno)
	}

	// Get the actual interface name
	var ifName [16]byte
	nameLen := uintptr(len(ifName))
	_, _, errno = syscall.Syscall6(syscall.SYS_GETSOCKOPT, uintptr(fd), 2, 2, uintptr(unsafe.Pointer(&ifName)), uintptr(unsafe.Pointer(&nameLen)), 0)
	if errno != 0 {
		syscall.Close(fd)
		return fmt.Errorf("failed to get interface name: %v", errno)
	}

	d.fd = fd
	d.name = string(ifName[:nameLen-1]) // Remove null terminator
	return nil
}