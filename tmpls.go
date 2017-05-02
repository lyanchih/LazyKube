package lazy

import (
	"encoding/json"
	"os"
	"text/template"
)

const OS_INSTALL_TMPL = `{
  "id": "coreos-install",
  "name": "CoreOS Install",
  "profile": "install-reboot",
  "metadata": {
    "coreos_channel": "{{.Channel}}",
    "coreos_version": "{{.Version}}",
    "ignition_endpoint": "{{.M.URL}}/ignition",
    "baseurl": "{{.M.URL}}/assets/coreos"
  }
}
`

const K8S_CONTROLLER_TMPL = `{
  "id": "{{.ID}}",
  "name": "k8s controller",
  "profile": "k8s-controller",
  "selector": {
    "mac": "{{index .MAC 0}}",
    "os": "installed"
  },
  "metadata": {
    "container_runtime": "docker",
    "domain_name": "{{.Domain}}",
    "etcd_initial_cluster": "{{.InitialCluster}}",
    "etcd_name": "{{.ID}}",
    "k8s_cert_endpoint": "{{.M.URL}}/assets",
    "k8s_dns_service_ip": "10.3.0.10",
    "k8s_etcd_endpoints": "{{.Endpoints}}",
    "k8s_pod_network": "10.2.0.0/16",
    "k8s_service_ip_range": "10.3.0.0/24",
    "vip": {{with .VIP}}{{ . }}{{ end }},
    "interfaces": {{.Nics}},
    "ssh_authorized_keys": {{.AuthorizedKeys}}
  }
}
`

const K8S_WORKER_TMPL = `{
  "id": "{{.ID}}",
  "name": "k8s worker",
  "profile": "k8s-worker",
  "selector": {
    "mac": "{{index .MAC 0}}",
    "os": "installed"
  },
  "metadata": {
    "container_runtime": "docker",
    "domain_name": "{{.Domain}}",
    "etcd_initial_cluster": "{{.InitialCluster}}",
    "k8s_controller_endpoint": "{{.ControllerEndpoint}}",
    "k8s_cert_endpoint": "{{.M.URL}}/assets",
    "k8s_dns_service_ip": "10.3.0.10",
    "k8s_etcd_endpoints": "{{.Endpoints}}",
    "interfaces": {{.Nics}},
    {{- with .Registries }}
    "registries": {{- j2s .}},
    {{- end }}
    "ssh_authorized_keys": {{.AuthorizedKeys}}
  }
}
`

const NODE_TMPL = `{
  "id": "{{.ID}}",
  "name": "Node {{.ID}}",
  "profile": "{{.Profile}}",
  "selector": {
    "mac": "{{index .MAC 0}}"
  },
  "metadata": {
    "domain_name": "{{.Domain}}",
    "interfaces": {{.Nics}},
    "ssh_authorized_keys": {{.AuthorizedKeys}}
  }
}
`

const DNSMASQ_TMPL = `# dnsmasq.conf

### DHCP CONFIG ###
{{- if .N.Gateway }}
dhcp-option=3,{{.N.Gateway}}
{{- end }}

{{- with (.Cls.GetKeepIPRange) }}
dhcp-range={{.Start}},{{.End}}
{{- end }}

{{- range $i, $node := .Nodes }}
  {{- range $index, $mac := .MAC }}
    {{- if lt $index (len $node.IP) }}
dhcp-host={{$mac}},{{index $node.IP $index}},1h
    {{- end }}
  {{- end }}
{{- end }}

dhcp-userclass=set:ipxe,iPXE
dhcp-boot=tag:#ipxe,undionly.kpxe
dhcp-boot=tag:ipxe,{{.M.URL}}/boot.ipxe

### TFTP CONFIG ###
enable-tftp
tftp-root=/var/lib/tftpboot

### DNS CONFIG ###

address=/{{.M.Domain}}/{{.M.IP}}

##### vip address #####
{{- with .V }}
  {{- if .Enable }}
address=/{{.Domain}}/{{.VIP}}
  {{- end }}
{{- end }}

##### node address #####
{{- range .Nodes }}
  {{- if gt (len .IP) 0 }}
address=/{{.Domain}}/{{index .IP 0}}
  {{- end }}
{{- end }}

##### dns server #####
{{- with .D }}
  {{- range .DNS }}
server={{.}}
  {{- end }}
{{- end }}

### OTHER CONFIG ###
log-queries
log-dhcp
`

var funcMap = template.FuncMap{
	"j2s": func(v interface{}) string {
		bs, _ := json.Marshal(v)
		return string(bs)
	},
}

func writeTemplateToFile(tmplContent, name, fileName string, data interface{}) error {
	tmpl, err := template.New(name).Funcs(funcMap).Parse(tmplContent)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}
