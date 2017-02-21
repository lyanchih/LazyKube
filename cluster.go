package lazy

type Cluster struct {
	InitialCluster     string
	Endpoints          string
	ControllerEndpoint string
	Hosts              map[string]string
	AuthorizedKeys     string
	M                  *MatchboxConfig
	*Network
}
