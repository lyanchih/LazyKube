package lazy

import (
	"encoding/json"
	"fmt"
)

type Node struct {
	*NodeConfig
	*Cluster
	ID     string
	Domain string
	Nics   NodeInterfaces
	VIP    *NodeInterface
}

type NodeInterfaces []NodeInterface

func (node *Node) makeInterfaces(c *Config) (NodeInterfaces, error) {
	nics := make(NodeInterfaces, 0, len(node.IP))
	dhcpChose := false
	for i, mac := range node.MAC {
		var ip string
		if i < len(node.IP) && len(node.IP[i]) != 0 &&
			c.Cls.ContainIP(node.IP[i], i) {
			ip = node.IP[i]
		}

		if len(ip) == 0 {
			ipnet, err := c.Cls.requestIP(mac, i)
			if err != nil {
				return nil, err
			}
			ip = ipnet.String()
			node.IP = append(node.IP, ip)
		}

		nic := NodeInterface{
			MAC:       mac,
			IP:        ip,
			Interface: fmt.Sprintf("%s%d", c.N.InterfaceBase, i),
		}

		if len(c.N.Gateway) != 0 && c.Cls.ContainIP(c.N.Gateway, i) {
			nic.Gateway = c.N.Gateway
		}

		if c.DHCP.Enable && !dhcpChose {
			if len(c.DHCP.Interface) == 0 || nic.Interface == c.DHCP.Interface {
				nic.DHCP = true
				dhcpChose = true
			}
		}

		nics = append(nics, nic)
	}
	return nics, nil
}

func (nis NodeInterfaces) String() string {
	bs, _ := json.Marshal(nis)
	return string(bs)
}

type NodeInterface struct {
	MAC       string `json:"mac"`
	IP        string `json:"ip"`
	Interface string `json:"interface"`
	DHCP      bool   `json:"dhcp"`
	Gateway   string `json:"gateway"`
}

func (nic NodeInterface) String() string {
	bs, _ := json.Marshal(nic)
	return string(bs)
}
