package localnetwork

import (
	"errors"
	"net"
	"reflect"
	"testing"
)

func TestDiscovererReturnsSortedPrivateIPv4URLsFromPhysicalInterfaces(t *testing.T) {
	discoverer := &Discoverer{
		interfaces: func() ([]net.Interface, error) {
			return []net.Interface{
				{Index: 1, Name: "en0", Flags: net.FlagUp},
				{Index: 2, Name: "docker0", Flags: net.FlagUp},
				{Index: 3, Name: "lo0", Flags: net.FlagUp | net.FlagLoopback},
			}, nil
		},
		addresses: func(networkInterface net.Interface) ([]net.Addr, error) {
			switch networkInterface.Name {
			case "en0":
				return []net.Addr{
					&net.IPNet{IP: net.ParseIP("192.168.1.42"), Mask: net.CIDRMask(24, 32)},
					&net.IPNet{IP: net.ParseIP("10.0.0.12"), Mask: net.CIDRMask(24, 32)},
					&net.IPNet{IP: net.ParseIP("8.8.8.8"), Mask: net.CIDRMask(24, 32)},
					&net.IPNet{IP: net.ParseIP("fd00::42"), Mask: net.CIDRMask(64, 128)},
				}, nil
			case "docker0":
				return []net.Addr{&net.IPNet{IP: net.ParseIP("172.17.0.1"), Mask: net.CIDRMask(16, 32)}}, nil
			default:
				return nil, nil
			}
		},
	}

	want := []string{"http://10.0.0.12:8096", "http://192.168.1.42:8096"}
	if got := discoverer.HTTPURLs(8096); !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected local URLs %#v", got)
	}
}

func TestDiscovererReturnsNoURLsWhenEnumerationFails(t *testing.T) {
	discoverer := &Discoverer{
		interfaces: func() ([]net.Interface, error) { return nil, errors.New("unavailable") },
	}
	if urls := discoverer.HTTPURLs(8096); len(urls) != 0 {
		t.Fatalf("expected no URLs, got %#v", urls)
	}
}

func TestDiscovererRejectsInvalidPort(t *testing.T) {
	if urls := NewDiscoverer().HTTPURLs(0); len(urls) != 0 {
		t.Fatalf("expected no URLs, got %#v", urls)
	}
}
