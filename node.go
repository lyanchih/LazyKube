package lazy

type Node struct {
	*NodeConfig
	*Cluster
	ID     string
	Domain string
}
