package lazy

type Cluster struct {
	InitialCluster     string
	Endpoints          string
	ControllerEndpoint string
	AuthorizedKeys     string
	Registries         []string
	M                  *MatchboxConfig
	*Network
}
