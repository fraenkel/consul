package consul

import (
	"fmt"
	"github.com/hashicorp/serf/serf"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/*
 * Contains an entry for each private block:
 * 10.0.0.0/8
 * 172.16.0.0/12
 * 192.168/16
 */
var privateBlocks []*net.IPNet

func init() {
	// Add each private block
	privateBlocks = make([]*net.IPNet, 3)
	_, block, err := net.ParseCIDR("10.0.0.0/8")
	if err != nil {
		panic(fmt.Sprintf("Bad cidr. Got %v", err))
	}
	privateBlocks[0] = block

	_, block, err = net.ParseCIDR("172.16.0.0/12")
	if err != nil {
		panic(fmt.Sprintf("Bad cidr. Got %v", err))
	}
	privateBlocks[1] = block

	_, block, err = net.ParseCIDR("192.168.0.0/16")
	if err != nil {
		panic(fmt.Sprintf("Bad cidr. Got %v", err))
	}
	privateBlocks[2] = block
}

// strContains checks if a list contains a string
func strContains(l []string, s string) bool {
	for _, v := range l {
		if v == s {
			return true
		}
	}
	return false
}

// ensurePath is used to make sure a path exists
func ensurePath(path string, dir bool) error {
	if !dir {
		path = filepath.Dir(path)
	}
	return os.MkdirAll(path, 0755)
}

// Returns if a member is a consul server. Returns a bool,
// the data center, and the rpc port
func isConsulServer(m serf.Member) (bool, string, int) {
	role := m.Role
	if !strings.HasPrefix(role, "consul:") {
		return false, "", 0
	}

	parts := strings.SplitN(role, ":", 3)
	datacenter := parts[1]
	port_str := parts[2]
	port, err := strconv.Atoi(port_str)
	if err != nil {
		return false, "", 0
	}

	return true, datacenter, port
}

// Returns if the given IP is in a private block
func isPrivateIP(ip_str string) bool {
	ip := net.ParseIP(ip_str)
	for _, priv := range privateBlocks {
		if priv.Contains(ip) {
			return true
		}
	}
	return false
}

// GetPrivateIP is used to return the first private IP address
// associated with an interface on the machine
func GetPrivateIP() (*net.IPNet, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("Failed to get interface addresses: %v", err)
	}

	// Find private IPv4 address
	for _, addr := range addresses {
		ip, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if ip.IP.To4() == nil {
			continue
		}
		if !isPrivateIP(ip.IP.String()) {
			continue
		}
		return ip, nil
	}
	return nil, fmt.Errorf("No private IP address found")
}