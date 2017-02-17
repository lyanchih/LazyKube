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
		log.Println("Parse session", s, "fail")
		return nil, err
	}
	return v, nil
}

func (cfg *iniConfig) newConfig() (*Config, error) {
	v, err := cfg.newConfigFromSection("", &Config{})
	return v.(*Config), err
}

func (cfg *iniConfig) newNodes(ids []string) ([]*Node, error) {
	nodes := make([]*Node, 0, len(ids))
	for _, id := range ids {
		if len(id) == 0 {
			continue
		}

		n, err := cfg.newConfigFromSection(id, &Node{})
		if err != nil {
			log.Println("New node config", id, "failed:", err)
			continue
		}
		nodes = append(nodes, n.(*Node))
	}
	return nodes, nil
}

func (cfg *iniConfig) newNetworkConfig() (*networkConfig, error) {
	v, err := cfg.newConfigFromSection("network", &networkConfig{})
	return v.(*networkConfig), err
}

func (cfg *iniConfig) newMatchboxConfig() (*matchboxConfig, error) {
	v, err := cfg.newConfigFromSection("matchbox", &matchboxConfig{})
	return v.(*matchboxConfig), err
}

type networkConfig struct {
	Gateway   string `ini:"gateway"`
	IPs       string `ini:"ips"`
	VIP       string `ini:"vip"`
	EnableVIP bool   `ini:"enable_vip"`
	VIPDomain string `ini:"vip_domain"`
}

type Config struct {
	Version    string   `ini:"version"`
	Channel    string   `ini:"channel"`
	DomainBase string   `ini:"domain_base"`
	NodeIDs    []string `ini:"nodes"`
	Keys       []string `ini:"keys"`
	n          *networkConfig
	M          *matchboxConfig
	nodes      []*Node
	cls        *Cluster
}

func Load(file string) (*Config, error) {
	cfg, err := loadINIConfig(file)
	if err != nil {
		log.Println("Load ini config failed:", err)
		return nil, err
	}

	c, err := cfg.newConfig()
	if err != nil {
		log.Println("Load config failed:", err)
		return nil, err
	}

	c.n, _ = cfg.newNetworkConfig()
	c.M, _ = cfg.newMatchboxConfig()
	c.nodes, _ = cfg.newNodes(c.NodeIDs)
	c.cls = &Cluster{
		M: c.M,
		Network: &Network{
			gateway: c.n.Gateway,
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
	for i, node := range c.nodes {
		node.ID = c.NodeIDs[i]
		node.Domain = node.ID
		if len(c.DomainBase) == 0 {
			node.Domain = node.Domain + "." + c.DomainBase
		}
		node.Cluster = c.cls
	}
	return nil
}

func (c *Config) analyzeCluster() error {
	initialCluster := make([]string, 0, len(c.nodes))
	endpoints := make([]string, 0, len(c.nodes))
	hosts := make(map[string]string)
	var controllerEndpoint string
	for _, n := range c.nodes {
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
	c.cls.InitialCluster = strings.Join(initialCluster, ",")
	c.cls.Endpoints = strings.Join(endpoints, ",")
	c.cls.ControllerEndpoint = controllerEndpoint
	c.cls.Hosts = hosts
	c.cls.AuthorizedKeys = string(bs)
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

	for _, n := range c.nodes {
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
	return nil
}
