package lazy

import (
	"errors"
	"log"
	"net"
	"regexp"
)

const bytePattern = "(?:[1-9]?[[:digit:]]|1[[:digit:]]{2}|2[0-4][[:digit:]]|25[0-5])"
const ipPattern = "(?:" + bytePattern + "\\.){3}" + bytePattern
const netmaskPattern = "(?:[1-2]?[[:digit:]]|3[0-2])"
const ipPoolPattern = "^(?P<ip>" + ipPattern + ")(?:/(?P<netmask>" + netmaskPattern + "))?(?::(?P<startIP>" + ipPattern + ")(?:-(?P<endIP>" + ipPattern + "))?)?$"

var (
	ipPoolReg        = regexp.MustCompile(ipPoolPattern)
	ipPoolMatchError = errors.New("IP pool is not match")
	startIPNotInCIDR = errors.New("Start IP of pool is not in CIDR")
	endIPNotInCIDR   = errors.New("End IP of pool is not in CIDR")
	endIPTooSmall    = errors.New("End IP should bigger start IP")
	ipIsNotEnough    = errors.New("IP pool is empty")
)

func ipv4ToUint32(ip net.IP) (n uint32) {
	if ip == nil {
		return n
	}
	if ip = ip.To4(); ip == nil {
		return n
	}

	for i, b := range ip {
		s := uint32(b)
		n += (1 << (uint(8 * (3 - i)))) * s
	}

	return n
}

func uint32ToIPv4(n uint32) (ip net.IP) {
	bs := make([]byte, 4)
	for i, v := range []uint32{1 << 24, 1 << 16, 1 << 8, 1} {
		bs[i] = byte(n / v)
		n = n % v
	}
	return net.IPv4(bs[0], bs[1], bs[2], bs[3])
}

func cidrLastIP(cidr net.IPNet) (ip net.IP) {
	ones, bits := cidr.Mask.Size()
	return uint32ToIPv4(ipv4ToUint32(cidr.IP) + uint32((1<<uint32(bits-ones))-1))
}

type Network struct {
	*NetworkConfig
	pools []networkPool
}

func newNetwork(nc *NetworkConfig) (*Network, error) {
	n := &Network{
		NetworkConfig: nc,
		pools:         make([]networkPool, 0, len(nc.IPs)),
	}
	for _, pool := range nc.IPs {
		np, err := newNetworkPool(pool)
		if err == ipPoolMatchError {
			log.Println(pool, "is not match with pool pattern")
			continue
		}

		if err != nil {
			log.Println("Parse pool failed with: ", err)
		}

		n.pools = append(n.pools, np)
	}
	return n, nil
}

func (n *Network) requestIP(mac string, poolIndex int) (net.IP, error) {
	if poolIndex >= len(n.pools) {
		return nil, errors.New("Only support for " + string(len(n.pools)) + " pools")
	}

	return n.pools[poolIndex].requestIP(mac)
}

type networkPool struct {
	net.IPNet
	startIP     net.IP
	endIP       net.IP
	startUint32 uint32
	endUint32   uint32
	currentIP   uint32
	pools       map[uint32]bool
}

func newNetworkPool(pool string) (np networkPool, err error) {
	if !ipPoolReg.MatchString(pool) {
		return np, ipPoolMatchError
	}

	ss := ipPoolReg.FindAllStringSubmatch(pool, -1)[0]
	matchs := make(map[string]string)
	for i, name := range ipPoolReg.SubexpNames() {
		if len(name) == 0 || len(ss[i]) == 0 {
			continue
		}
		matchs[name] = ss[i]
	}

	var s string
	var ok bool
	if s, ok = matchs["ip"]; ok {
		np.IP = net.ParseIP(s)
		np.Mask = np.IP.DefaultMask()
		_, ipnet, _ := net.ParseCIDR(np.IPNet.String())
		np.IPNet = *ipnet
	}

	if s, ok = matchs["netmask"]; ok {
		_, ipnet, _ := net.ParseCIDR(np.IP.String() + "/" + string(s))
		np.IPNet = *ipnet
	}

	if s, ok = matchs["startIP"]; ok {
		np.startIP = net.ParseIP(s)
		if !np.Contains(np.startIP) {
			np.startIP = nil
			err = startIPNotInCIDR
		}
	}

	if s, ok = matchs["endIP"]; ok && np.startIP != nil && err == nil {
		np.endIP = net.ParseIP(s)
		if !np.Contains(np.endIP) {
			np.endIP = nil
			err = endIPNotInCIDR
		} else if ipv4ToUint32(np.endIP) < ipv4ToUint32(np.startIP) {
			np.endIP = nil
			err = endIPTooSmall
		}
	}

	if np.startIP == nil {
		np.startIP = uint32ToIPv4(ipv4ToUint32(np.IP) + 1)
	}

	if np.endIP == nil {
		np.endIP = cidrLastIP(np.IPNet)
	}

	np.startUint32 = ipv4ToUint32(np.startIP)
	np.endUint32 = ipv4ToUint32(np.endIP)
	np.currentIP = np.startUint32
	np.pools = make(map[uint32]bool)
	return np, err
}

func (np *networkPool) requestIP(mac string) (net.IP, error) {
	if np.currentIP > np.endUint32 {
		for uint32IP, used := range np.pools {
			if !used {
				return uint32ToIPv4(uint32IP), nil
			}
		}
		return nil, ipIsNotEnough
	}

	ip := uint32ToIPv4(np.currentIP)
	np.pools[np.currentIP] = true
	np.currentIP = np.currentIP + 1
	return ip, nil
}
