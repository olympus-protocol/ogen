package host

import (
	"crypto/rand"
	"fmt"
	"github.com/libp2p/go-libp2p"
	circuit "github.com/libp2p/go-libp2p-circuit"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"sort"
	"time"
)

// Retrieves an external ipv4 address and converts into a libp2p formatted value.
func ipAddr() net.IP {
	ip, err := ExternalIP()
	if err != nil {
		log.Fatalf("Could not get IPv4 address: %v", err)
	}
	return net.ParseIP(ip)
}

// ExternalIPv4 returns the first IPv4 available.
func ExternalIPv4() (string, error) {
	ips, err := retrieveIPAddrs()
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "127.0.0.1", nil
	}
	for _, ip := range ips {
		ip = ip.To4()
		if ip == nil {
			continue // not an ipv4 address
		}
		return ip.String(), nil
	}
	return "127.0.0.1", nil
}

// ExternalIPv6 retrieves any allocated IPv6 addresses
// from the accessible network interfaces.
func ExternalIPv6() (string, error) {
	ips, err := retrieveIPAddrs()
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "127.0.0.1", nil
	}
	for _, ip := range ips {
		if ip.To4() != nil {
			continue // not an ipv6 address
		}
		if ip.To16() == nil {
			continue
		}
		return ip.String(), nil
	}
	return "127.0.0.1", nil
}

// ExternalIP returns the first IPv4/IPv6 available.
func ExternalIP() (string, error) {
	ips, err := retrieveIPAddrs()
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "127.0.0.1", nil
	}
	return ips[0].String(), nil
}

func retrieveIPAddrs() ([]net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var ipAddrs []net.IP
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}
			ipAddrs = append(ipAddrs, ip)
		}
	}
	return SortAddresses(ipAddrs), nil
}

// SortAddresses sorts a set of addresses in the order of
// ipv4 -> ipv6.
func SortAddresses(ipAddrs []net.IP) []net.IP {
	sort.Slice(ipAddrs, func(i, j int) bool {
		return ipAddrs[i].To4() != nil && ipAddrs[j].To4() == nil
	})
	return ipAddrs
}

func multiAddressBuilder(ipAddr string, port string) (ma.Multiaddr, error) {
	parsedIP := net.ParseIP(ipAddr)
	if parsedIP.To4() == nil && parsedIP.To16() == nil {
		return nil, errors.Errorf("invalid ip address provided: %s", ipAddr)
	}
	if parsedIP.To4() != nil {
		return ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", ipAddr, port))
	}
	return ma.NewMultiaddr(fmt.Sprintf("/ip6/%s/tcp/%s", ipAddr, port))
}

func multiAddressBuilderWithID(ipAddr, protocol string, port string, id peer.ID) (ma.Multiaddr, error) {
	parsedIP := net.ParseIP(ipAddr)
	if parsedIP.To4() == nil && parsedIP.To16() == nil {
		return nil, errors.Errorf("invalid ip address provided: %s", ipAddr)
	}
	if id.String() == "" {
		return nil, errors.New("empty peer id given")
	}
	if parsedIP.To4() != nil {
		return ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/%s/%s/p2p/%s", ipAddr, protocol, port, id.String()))
	}
	return ma.NewMultiaddr(fmt.Sprintf("/ip6/%s/%s/%s/p2p/%s", ipAddr, protocol, port, id.String()))
}

func buildOptions(priKey crypto.PrivKey, ps peerstore.Peerstore) []libp2p.Option {
	netParams := config.GlobalParams.NetParams
	var listen []ma.Multiaddr
	var ips []string
	ipv4, err := ExternalIPv4()
	if err == nil {
		ips = append(ips, ipv4)
	}
	ipv6, err := ExternalIPv6()
	if err == nil {
		ips = append(ips, ipv6)
	}
	if len(ips) == 0 {
		panic("no ip available to listen")
	}
	for _, ip := range ips {
		addr, err := multiAddressBuilder(ip, netParams.DefaultP2PPort)
		if err != nil {
			panic("no ip available to listen")
		}
		listen = append(listen, addr)
	}

	connman := connmgr.NewConnManager(2, 64, time.Second*60)

	options := []libp2p.Option{
		libp2p.Identity(priKey),
		libp2p.ListenAddrs(listen...),
		libp2p.NATPortMap(),
		libp2p.UserAgent(fmt.Sprintf("ogen/%s", params.Version)),
		libp2p.ConnectionManager(connman),
		libp2p.Peerstore(ps),
		libp2p.EnableRelay(circuit.OptActive, circuit.OptHop),
	}

	return options
}

func (h *host) loadPrivateKey() (crypto.PrivKey, error) {
	keyBytes, err := ioutil.ReadFile(path.Join(h.datapath, "node_key.dat"))
	if err != nil {
		return h.createPrivateKey()
	}

	key, err := crypto.UnmarshalPrivateKey(keyBytes)
	if err != nil {
		return h.createPrivateKey()
	}
	return key, nil
}

func (h *host) createPrivateKey() (crypto.PrivKey, error) {
	_ = os.RemoveAll(path.Join(h.datapath, "node_key.dat"))

	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}

	keyBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(path.Join(h.datapath, "node_key.dat"), keyBytes, 0700)
	if err != nil {
		return nil, err
	}

	return priv, nil
}
