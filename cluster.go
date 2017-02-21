package lazy

type Cluster struct {
	InitialCluster     string
	Endpoints          string
	ControllerEndpoint string
	AuthorizedKeys     string
	M                  *MatchboxConfig
	*Network
}
