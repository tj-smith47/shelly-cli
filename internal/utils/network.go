package utils

import (
	"fmt"
	"net"
)

// DetectSubnet attempts to detect the local network subnet.
// It returns a CIDR notation string like "192.168.1.0/24".
func DetectSubnet() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				// Return the network address with mask
				network := ipNet.IP.Mask(ipNet.Mask)
				ones, _ := ipNet.Mask.Size()
				return fmt.Sprintf("%s/%d", network, ones), nil
			}
		}
	}

	return "", fmt.Errorf("no suitable network interface found")
}
