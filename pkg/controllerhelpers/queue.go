package controllerhelpers

type ClusterKey struct{}

func (k ClusterKey) ObjectName() string {
	return "cluster"
}

func MakeClusterKey() ClusterKey {
	return ClusterKey{}
}
