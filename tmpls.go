package lazy

import (
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
    "mac": "{{.MAC}}",
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
    "ssh_authorized_keys": {{.AuthorizedKeys}}
  }
}
`

const K8S_WORKER_TMPL = `{
  "id": "{{.ID}}",
  "name": "k8s worker",
  "profile": "k8s-worker",
  "selector": {
    "mac": "{{.MAC}}",
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
    "ssh_authorized_keys": {{.AuthorizedKeys}}
  }
}
`

func writeTemplateToFile(tmplContent, name, fileName string, data interface{}) error {
	tmpl, err := template.New(name).Parse(tmplContent)
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
