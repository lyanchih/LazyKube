package lazy

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-ini/ini"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type iniConfig ini.File

func loadINIConfig(file string) (*iniConfig, error) {
	cfg, err := ini.Load(file)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	cfg.NameMapper = ini.TitleUnderscore
	return (*iniConfig)(cfg), nil
}

func (cfg *iniConfig) newConfigFromSection(s string, v interface{}) (interface{}, error) {
	iniFile := (*ini.File)(cfg)
	sec, err := iniFile.GetSection(s)
	if err != nil {
		return v, nil
	}

	err = sec.MapTo(v)
	if err != nil {
		if len(s) == 0 {
			s = "DEFAULT"
		}
		log.Println("Parse session", s, "fail")
		return nil, err
	}
	return v, nil
}

func (cfg *iniConfig) newDefaultConfig() (*DefaultConfig, error) {
	v, err := cfg.newConfigFromSection("", &DefaultConfig{})
	if err != nil {
		return nil, err
	}
	return v.(*DefaultConfig), nil
}

func (cfg *iniConfig) newNodes(ids []string) ([]*Node, error) {
	nodes := make([]*Node, 0, len(ids))
	for _, id := range ids {
		if len(id) == 0 {
			continue
		}

		n, err := cfg.newConfigFromSection(id, &NodeConfig{})
		if err != nil {
			log.Println("New node config", id, "failed:", err)
			continue
		}
		nodes = append(nodes, &Node{NodeConfig: n.(*NodeConfig)})
	}
	return nodes, nil
}

func (cfg *iniConfig) newContainerConfig() (*ContainerConfig, error) {
	v, err := cfg.newConfigFromSection("container", &ContainerConfig{})
	if err != nil {
		return nil, err
	}

	n := v.(*ContainerConfig)
	return n, nil
}

func (cfg *iniConfig) newNetworkConfig() (*NetworkConfig, error) {
	v, err := cfg.newConfigFromSection("network", &NetworkConfig{})
	if err != nil {
		return nil, err
	}

	n := v.(*NetworkConfig)
	if len(n.InterfaceBase) == 0 {
		n.InterfaceBase = "eth"
	}
	return n, nil
}

func (cfg *iniConfig) newMatchboxConfig() (*MatchboxConfig, error) {
	v, err := cfg.newConfigFromSection("matchbox", &MatchboxConfig{})
	if err != nil {
		return nil, err
	}
	return v.(*MatchboxConfig), nil
}

func (cfg *iniConfig) newDNSConfig() (*DNSConfig, error) {
	v, err := cfg.newConfigFromSection("dns", &DNSConfig{})
	if err != nil {
		return nil, err
	}
	return v.(*DNSConfig), nil
}

func (cfg *iniConfig) newDHCPConfig() (*DHCPConfig, error) {
	v, err := cfg.newConfigFromSection("dns", &DHCPConfig{})
	if err != nil {
		return nil, err
	}

	return v.(*DHCPConfig), nil
}

func (cfg *iniConfig) newVIPConfig() (*VIPConfig, error) {
	v, err := cfg.newConfigFromSection("vip", &VIPConfig{})
	if err != nil {
		return nil, err
	}
	return v.(*VIPConfig), nil
}

type Config struct {
	*DefaultConfig
	C     *ContainerConfig
	N     *NetworkConfig
	M     *MatchboxConfig
	D     *DNSConfig
	DHCP  *DHCPConfig
	V     *VIPConfig
	Nodes []*Node
	Cls   *Cluster
}

type DefaultConfig struct {
	Version    string   `ini:"version"`
	Channel    string   `ini:"channel"`
	DomainBase string   `ini:"domain_base"`
	NodeIDs    []string `ini:"nodes"`
	Keys       []string `ini:"keys"`
}

type MatchboxConfig struct {
	URL    string `ini:"url"`
	IP     string `ini:"ip"`
	Domain string `ini:"domain"`
}

type NodeConfig struct {
	MAC     []string `ini:"mac"`
	Role    string   `ini:"role"`
	IP      []string `ini:"ip"`
	Profile string   `ini:"profile"`
}

type ContainerConfig struct {
	Registries []string `ini:"registries"`
}

type NetworkConfig struct {
	Gateway       string   `ini:"gateway"`
	IPs           []string `ini:"ips"`
	DHCP_keep     int      `ini:"dhcp_keep"`
	InterfaceBase string   `ini:"interface_base"`
}

type DNSConfig struct {
	DNS    []string `ini:"dns"`
	Driver string   `ini:"driver"`
}

type DHCPConfig struct {
	Enable    bool   `ini:"enable"`
	Interface string `ini:"interface"`
}

type VIPConfig struct {
	Enable bool   `ini:"enable"`
	VIP    string `ini:"vip"`
	Domain string `ini:"domain"`
}

func Load(file string) (*Config, error) {
	cfg, err := loadINIConfig(file)
	if err != nil {
		log.Println("Load ini config failed:", err)
		return nil, err
	}

	c := &Config{}

	if c.DefaultConfig, err = cfg.newDefaultConfig(); err != nil {
		log.Println("Load config failed:", err)
		return nil, err
	}

	if c.C, err = cfg.newContainerConfig(); err != nil {
		log.Println("Load container config failed:", err)
		return nil, err
	}

	if c.N, err = cfg.newNetworkConfig(); err != nil {
		log.Println("Load network config failed:", err)
		return nil, err
	}

	if c.M, err = cfg.newMatchboxConfig(); err != nil {
		log.Println("Load matchbox config failed:", err)
		return nil, err
	}

	if c.D, err = cfg.newDNSConfig(); err != nil {
		log.Println("Load dns config failed:", err)
		return nil, err
	}

	if c.DHCP, err = cfg.newDHCPConfig(); err != nil {
		log.Println("Load dhcp config failed:", err)
		return nil, err
	}

	if c.V, err = cfg.newVIPConfig(); err != nil {
		log.Println("Load vip config failed:", err)
		return nil, err
	}

	if c.Nodes, err = cfg.newNodes(c.NodeIDs); err != nil {
		log.Println("Load nodes failed:", err)
		return nil, err
	}

	c.Cls = &Cluster{
		M: c.M,
	}

	if err = c.analyze(); err != nil {
		log.Println(err)
		return nil, err
	}
	return c, nil
}

func (c *Config) analyze() (err error) {
	if err = c.analyzeNetwork(); err != nil {
		return errors.New("Analyze network failed: " + err.Error())
	}
	if err = c.analyzeMatchbox(); err != nil {
		return errors.New("Analyze matchbox failed: " + err.Error())
	}
	if err = c.analyzeNodes(); err != nil {
		return errors.New("Analyze nodes failed: " + err.Error())
	}
	if err = c.analyzeVIP(); err != nil {
		return errors.New("Analyze VIP failed: " + err.Error())
	}
	if err = c.analyzeCluster(); err != nil {
		return errors.New("Analyze cluster failed: " + err.Error())
	}
	return nil
}

func (c *Config) analyzeNetwork() error {
	n, err := newNetwork(c.N)
	if err != nil {
		return err
	}

	c.Cls.Network = n
	return nil
}

func (c *Config) analyzeMatchbox() error {
	// TODO: Implement network
	return nil
}

func (c *Config) analyzeNodes() error {
	for i, node := range c.Nodes {
		node.ID = c.NodeIDs[i]
		node.Domain = node.ID
		if len(c.DomainBase) != 0 {
			node.Domain = node.Domain + "." + c.DomainBase
		}

		if len(node.Profile) == 0 {
			node.Profile = "node"
		}

		nics, err := node.makeInterfaces(c)
		if err != nil {
			log.Println("Make interfaces failed: ", err)
			return err
		}
		node.Nics = nics
		node.Cluster = c.Cls
	}
	return nil
}

func (c *Config) analyzeVIP() error {
	if c.V == nil || !c.V.Enable {
		return nil
	}

	vip := c.V.VIP
	if !validateIPv4(vip) {
		return errors.New("VIP format is not correct: " + vip)
	}

	for _, np := range c.Cls.Network.pools {
		if !np.Contains(net.ParseIP(vip)) {
			continue
		}

		for _, node := range c.Nodes {
			for _, nic := range node.Nics {
				if !np.Contains(net.ParseIP(nic.IP)) {
					continue
				}

				node.VIP = &NodeInterface{
					IP:        vip,
					Interface: nic.Interface,
				}
			}
		}
		return nil
	}

	return errors.New("VIP can not find relative network pool: " + vip)
}

func (c *Config) analyzeCluster() error {
	var controllerEndpoint string
	initialCluster := make([]string, 0, len(c.Nodes))
	endpoints := make([]string, 0, len(c.Nodes))

	for _, n := range c.Nodes {
		if n.Role == "master" {
			initialCluster = append(initialCluster, fmt.Sprintf("%s=http://%s:2380", n.ID, n.Domain))
			endpoints = append(endpoints, fmt.Sprintf("http://%s:2379", n.Domain))
			if len(controllerEndpoint) == 0 {
				controllerEndpoint = fmt.Sprintf("https://%s", n.Domain)
			}
		}
	}

	if c.V != nil && c.V.Enable {
		controllerEndpoint = fmt.Sprintf("https://%s", c.V.Domain)
	}

	bs, err := json.Marshal(c.Keys)
	if err != nil {
		log.Println("Convert keys into json array failed: ", err)
		bs = []byte("[]")
	}

	c.Cls.InitialCluster = strings.Join(initialCluster, ",")
	c.Cls.Endpoints = strings.Join(endpoints, ",")
	c.Cls.ControllerEndpoint = controllerEndpoint
	c.Cls.AuthorizedKeys = string(bs)
	c.Cls.Registries = c.C.Registries
	return nil
}

func (c *Config) Generate(outputPath string) error {
	err := os.MkdirAll(outputPath, 0744)
	if err != nil {
		log.Fatal("Make output path ", outputPath, "fail")
	}
	err = writeTemplateToFile(OS_INSTALL_TMPL, "install",
		filepath.Join(outputPath, "install.json"), c)
	if err != nil {
		log.Println("Write template install failed: ", err)
	}

	for _, n := range c.Nodes {
		var tmpl, name string
		switch n.Role {
		case "master":
			tmpl = K8S_CONTROLLER_TMPL
			name = "controller"
		case "minion":
			tmpl = K8S_WORKER_TMPL
			name = "worker"
		default:
			tmpl = NODE_TMPL
			name = "Node " + n.ID
		}

		err = writeTemplateToFile(tmpl, name,
			filepath.Join(outputPath, n.ID+".json"), n)
		if err != nil {
			log.Println("Write template install failed: ", err)
		}
	}

	err = writeTemplateToFile(DNSMASQ_TMPL, "dnsmasq",
		filepath.Join(outputPath, "dnsmasq.conf"), c)
	if err != nil {
		log.Println("Write dnsmasq config failed: ", err)
	}
	return nil
}
