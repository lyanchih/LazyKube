package lazy

import (
	"encoding/json"
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

func (nis NodeInterfaces) String() string {
	bs, _ := json.Marshal(nis)
	return string(bs)
}

type NodeInterface struct {
	MAC       string `json:"mac"`
	IP        string `json:"ip"`
	Interface string `json:"interface"`
}

func (nic NodeInterface) String() string {
	bs, _ := json.Marshal(nic)
	return string(bs)
}
