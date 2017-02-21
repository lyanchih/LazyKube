package lazy

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/go-ini/ini"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var outputPath string

func init() {
	flag.StringVar(&outputPath, "-output", "_output", "Output path")
}

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

func (cfg *iniConfig) newNetworkConfig() (*NetworkConfig, error) {
	v, err := cfg.newConfigFromSection("network", &NetworkConfig{})
	if err != nil {
		return nil, err
	}
	return v.(*NetworkConfig), nil
}

func (cfg *iniConfig) newMatchboxConfig() (*MatchboxConfig, error) {
	v, err := cfg.newConfigFromSection("matchbox", &MatchboxConfig{})
	if err != nil {
		return nil, err
	}
	return v.(*MatchboxConfig), nil
}

type Config struct {
	*DefaultConfig
	N     *NetworkConfig
	M     *MatchboxConfig
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
	MAC  string `ini:"mac"`
	Role string `ini:"role"`
	IP   string `ini:"ip"`
}

type NetworkConfig struct {
	Gateway   string   `ini:"gateway"`
	IPs       string   `ini:"ips"`
	VIP       string   `ini:"vip"`
	DNS       []string `ini:"dns"`
	EnableVIP bool     `ini:"enable_vip"`
	VIPDomain string   `ini:"vip_domain"`
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

	if c.N, err = cfg.newNetworkConfig(); err != nil {
		log.Println("Load network config failed:", err)
	}

	if c.M, err = cfg.newMatchboxConfig(); err != nil {
		log.Println("Load matchbox config failed:", err)
	}

	if c.Nodes, err = cfg.newNodes(c.NodeIDs); err != nil {
		log.Println("Load nodes failed:", err)
	}

	c.Cls = &Cluster{
		M: c.M,
		Network: &Network{
			NetworkConfig: c.N,
		},
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
	if err = c.analyzeCluster(); err != nil {
		return errors.New("Analyze cluster failed: " + err.Error())
	}
	return nil
}

func (c *Config) analyzeNetwork() error {
	// TODO: Implement network
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
		if len(c.DomainBase) == 0 {
			node.Domain = node.Domain + "." + c.DomainBase
		}
		node.Cluster = c.Cls
	}
	return nil
}

func (c *Config) analyzeCluster() error {
	initialCluster := make([]string, 0, len(c.Nodes))
	endpoints := make([]string, 0, len(c.Nodes))
	hosts := make(map[string]string)
	var controllerEndpoint string
	for _, n := range c.Nodes {
		if n.Role == "master" {
			initialCluster = append(initialCluster, fmt.Sprintf("%s=http://%s:2380", n.ID, n.Domain))
			endpoints = append(endpoints, fmt.Sprintf("http://%s:2379", n.Domain))
			if len(controllerEndpoint) == 0 {
				controllerEndpoint = fmt.Sprintf("https://%s", n.Domain)
			}
		}
		hosts[n.IP] = n.Domain
	}
	bs, err := json.Marshal(c.Keys)
	if err != nil {
		log.Println("Convert keys into json array failed: ", err)
		bs = []byte("[]")
	}
	c.Cls.InitialCluster = strings.Join(initialCluster, ",")
	c.Cls.Endpoints = strings.Join(endpoints, ",")
	c.Cls.ControllerEndpoint = controllerEndpoint
	c.Cls.Hosts = hosts
	c.Cls.AuthorizedKeys = string(bs)
	return nil
}

func (c *Config) Generate() error {
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
