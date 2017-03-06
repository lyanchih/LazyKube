package lazy

import (
	"net"
	"regexp"
	"testing"
)

func testRegMatch(t *testing.T, r *regexp.Regexp, in []string, b bool, category string) {
	for _, s := range in {
		if r.MatchString(s) != b {
			t.Fatalf("%s reg is not match with %s, it should be %v\n", category, s, b)
		}
	}
}

func TestByteReg(t *testing.T) {
	r, err := regexp.Compile("^" + bytePattern + "$")
	if err != nil {
		t.Fatal(err)
	}

	category := "Byte"
	in := []string{"1", "10", "123", "210", "251"}
	testRegMatch(t, r, in, true, category)

	in = []string{"01", "256", "299", "399", "599", "1234"}
	testRegMatch(t, r, in, false, category)
}

func TestIPPatternReg(t *testing.T) {
	r, err := regexp.Compile("^" + ipPattern + "$")
	if err != nil {
		t.Fatal(err)
	}

	category := "IP"
	in := []string{"1.2.3.4", "172.17.0.100"}
	testRegMatch(t, r, in, true, category)

	in = []string{"1.2.3.256", "8.8.8+8", "8.8+8.8", "1.2.3.", "1.2.", "256.1.2.3"}
	testRegMatch(t, r, in, false, category)
}

func TestNetmaskPatterReg(t *testing.T) {
	r, err := regexp.Compile("^" + netmaskPattern + "$")
	if err != nil {
		t.Fatal(err)
	}

	category := "Netmask"
	in := []string{"0", "1", "9", "12", "24", "32"}
	testRegMatch(t, r, in, true, category)

	in = []string{"-5", "33", "40"}
	testRegMatch(t, r, in, false, category)
}

func TestIPPoolPatternReg(t *testing.T) {
	r, err := regexp.Compile(ipPoolPattern)
	if err != nil {
		t.Fatal(err)
	}

	category := "IPs"
	in := []string{"192.168.0.0", "192.168.0.0/24", "192.168.0.0/24:192.168.0.10",
		"192.168.0.0/24:192.168.0.10-192.168.0.20", "192.168.0.0/24:192.168.0.10-192.168.0.250"}
	testRegMatch(t, r, in, true, category)

	in = []string{"192.168.0.256", "192.168.0.0\\24", "192.168.0.0/24|192.168.0.10",
		"192.168.0.0/24-192.168.0.10", "192.168.0.0/24-192.168.0.10", "192.168.0.0/24:-192.168.0.10"}
	testRegMatch(t, r, in, false, category)
}

func TestIPv4ToUint32ToIPv4(t *testing.T) {
	testFunc := func(ip string, n uint32) {
		if ipv4ToUint32(net.ParseIP(ip)) != n {
			t.Fatalf("%s shoud be convert to %n", ip, n)
		}

		if uint32ToIPv4(n).String() != ip {
			t.Fatalf("%n shoud be convert to %s", n, ip)
		}
	}

	testFunc("122.116.113.134", 2054451590)
	testFunc("8.8.8.8", 134744072)
	testFunc("254.3.6.100", 4261611108)
	testFunc("10.20.30.40", 169090600)
	testFunc("172.32.200.5", 2887829509)
}

func TestCIDRLastIP(t *testing.T) {
	testFunc := func(s string, ip string) {
		_, ipnet, _ := net.ParseCIDR(s)
		if cidrLastIP(*ipnet).String() != ip {
			t.Fatalf("%v IPnet last ip is not %s", ipnet, ip)
		}
	}

	testFunc("172.17.0.14/24", "172.17.0.255")
	testFunc("172.17.16.14/28", "172.17.16.15")
	testFunc("192.55.121.14/16", "192.55.255.255")
}

func testPoolResult(t *testing.T, p networkPool, ipnet, start, end string) {
	if p.IPNet.String() != ipnet {
		t.Fatalf("%v IPnet is not same with %s", p, ipnet)
	}

	if (p.startIP == nil && len(start) != 0) || (p.startIP != nil && p.startIP.String() != start) {
		t.Fatalf("%v start IP is not same with %s", p, start)
	}

	if (p.endIP == nil && len(end) != 0) || (p.endIP != nil && p.endIP.String() != end) {
		t.Fatalf("%v end IP is not same with %s", p, end)
	}
}

func TestNewNetworkPool(t *testing.T) {
	// Test default class A netmask
	p, err := newNetworkPool("8.0.103.5", ipPoolKeepIP)
	if err != nil {
		t.Fatal(err)
	}
	testPoolResult(t, p, "8.0.0.0/8", "8.0.0.1", "8.255.255.255")

	// Test default class B netmask
	p, err = newNetworkPool("172.32.200.5", ipPoolKeepIP)
	if err != nil {
		t.Fatal(err)
	}
	testPoolResult(t, p, "172.32.0.0/16", "172.32.0.1", "172.32.255.255")

	// Test default class C netmask
	p, err = newNetworkPool("192.168.0.5", ipPoolKeepIP)
	if err != nil {
		t.Fatal(err)
	}
	testPoolResult(t, p, "192.168.0.0/24", "192.168.0.1", "192.168.0.255")

	// Test custom netmask
	p, err = newNetworkPool("192.168.0.5/14", ipPoolKeepIP)
	if err != nil {
		t.Fatal(err)
	}
	testPoolResult(t, p, "192.168.0.0/14", "192.168.0.1", "192.171.255.255")

	// Test custom netmask with start ip
	p, err = newNetworkPool("192.168.0.5/16:192.168.250.87", ipPoolKeepIP)
	if err != nil {
		t.Fatal(err)
	}
	testPoolResult(t, p, "192.168.0.0/16", "192.168.250.87", "192.168.255.255")

	// Test custom netmask with end ip
	p, err = newNetworkPool("192.168.0.5/16:192.168.0.87-192.168.5.144", ipPoolKeepIP)
	if err != nil {
		t.Fatal(err)
	}
	testPoolResult(t, p, "192.168.0.0/16", "192.168.0.87", "192.168.5.144")

	// Test default netmask with wrong start ip
	s := "192.168.56.5:172.0.0.5"
	p, err = newNetworkPool(s, ipPoolKeepIP)
	if err != startIPNotInCIDR {
		t.Fatalf("%s should have startIPNotInCIDR error\n", s)
	}
	testPoolResult(t, p, "192.168.56.0/24", "192.168.56.1", "192.168.56.255")

	// Test netmask with wrong start ip
	s = "192.168.0.5/16:172.0.0.5"
	p, err = newNetworkPool(s, ipPoolKeepIP)
	if err != startIPNotInCIDR {
		t.Fatalf("%s should have startIPNotInCIDR error\n", s)
	}
	testPoolResult(t, p, "192.168.0.0/16", "192.168.0.1", "192.168.255.255")

	// Test default netmask with wrong end ip
	s = "192.168.0.5:192.168.0.2-192.168.2.5"
	p, err = newNetworkPool(s, ipPoolKeepIP)
	if err != endIPNotInCIDR {
		t.Fatalf("%s should have endIPNotInCIDR error\n", s)
	}
	testPoolResult(t, p, "192.168.0.0/24", "192.168.0.2", "192.168.0.255")

	// Test netmask with wrong end ip
	s = "192.168.0.5/16:192.168.128.2-192.169.2.5"
	p, err = newNetworkPool(s, ipPoolKeepIP)
	if err != endIPNotInCIDR {
		t.Fatalf("%s should have endIPNotInCIDR error\n", s)
	}
	testPoolResult(t, p, "192.168.0.0/16", "192.168.128.2", "192.168.255.255")

	// Test netmask with smaller end ip
	s = "192.168.0.5/16:192.168.128.10-192.168.128.1"
	p, err = newNetworkPool(s, ipPoolKeepIP)
	if err != endIPTooSmall {
		t.Fatalf("%s should have endIPTooSmall error\n", s)
	}
	testPoolResult(t, p, "192.168.0.0/16", "192.168.128.10", "192.168.255.255")
}

func TestNetworkPoolRequestIP(t *testing.T) {
	var s string
	var p networkPool
	var err error
	testRequest := func(ipString string, expectedErr error) {
		ip, err := p.requestIP("")
		if expectedErr != nil && err == expectedErr {
			return
		}
		if err != nil {
			t.Fatalf("%s request ip %s failed: %s", s, ipString, err)
		}

		if ipString != ip.String() {
			t.Fatalf("%s request %s is not match with %s", s, ip.String(), ipString)
		}
	}

	s = "192.168.99.5/24:192.168.99.55"
	p, err = newNetworkPool(s, ipPoolKeepIP)
	if err != nil {
		t.Fatal(err)
	}
	testRequest("192.168.99.55", nil)
	testRequest("192.168.99.56", nil)
	testRequest("192.168.99.57", nil)

	s = "1.2.3.5/16:1.2.3.105-1.2.3.127"
	p, err = newNetworkPool(s, ipPoolKeepIP)
	if err != nil {
		t.Fatal(err)
	}
	testRequest("1.2.3.105", nil)
	testRequest("1.2.3.106", nil)
	testRequest("1.2.3.107", nil)
	testRequest("", ipIsNotEnough)

	s = "1.2.3.5/16:1.2.3.200-1.2.3.202"
	p, err = newNetworkPool(s, 2)
	if err != nil {
		t.Fatal(err)
	}
	testRequest("1.2.3.200", nil)
	testRequest("", ipIsNotEnough)

	// test ip is not enought for keep ip
	s = "1.2.3.5/24:1.2.3.240"
	p, err = newNetworkPool(s, ipPoolKeepIP)
	if err != poolCanNotKeep {
		t.Fatal(err)
	}
}
