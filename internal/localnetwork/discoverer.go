package localnetwork

import (
	"net"
	"sort"
	"strconv"
	"strings"
)

type Discoverer struct {
	interfaces func() ([]net.Interface, error)
	addresses  func(net.Interface) ([]net.Addr, error)
}

func NewDiscoverer() *Discoverer {
	return &Discoverer{
		interfaces: net.Interfaces,
		addresses:  func(networkInterface net.Interface) ([]net.Addr, error) { return networkInterface.Addrs() },
	}
}

func (d *Discoverer) HTTPURLs(port int) []string {
	if port < 1 || port > 65535 {
		return nil
	}
	interfaces, err := d.interfaces()
	if err != nil {
		return nil
	}

	unique := make(map[string]struct{})
	for _, networkInterface := range interfaces {
		if !usableInterface(networkInterface) {
			continue
		}
		addresses, addressErr := d.addresses(networkInterface)
		if addressErr != nil {
			continue
		}
		for _, address := range addresses {
			ip, _, parseErr := net.ParseCIDR(address.String())
			if parseErr != nil || ip.To4() == nil || !ip.IsPrivate() {
				continue
			}
			unique["http://"+net.JoinHostPort(ip.String(), strconv.Itoa(port))] = struct{}{}
		}
	}

	urls := make([]string, 0, len(unique))
	for url := range unique {
		urls = append(urls, url)
	}
	sort.Strings(urls)
	return urls
}

func usableInterface(networkInterface net.Interface) bool {
	if networkInterface.Flags&net.FlagUp == 0 || networkInterface.Flags&net.FlagLoopback != 0 {
		return false
	}
	name := strings.ToLower(networkInterface.Name)
	for _, prefix := range []string{
		"awdl", "br-", "bridge", "docker", "llw", "podman", "utun", "veth", "vmnet",
	} {
		if strings.HasPrefix(name, prefix) {
			return false
		}
	}
	return true
}
