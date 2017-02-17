package lazy

type Node struct {
	MAC  string `ini:"mac"`
	Role string `ini:"role"`
	IP   string `ini:"ip"`

	ID     string
	Domain string
	*Cluster
}
