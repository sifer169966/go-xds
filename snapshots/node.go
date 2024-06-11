package snapshots

import corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

// DefaultNodeID ...
// uses the default nodeID to let all clients relies on the same version of resources
type DefaultNodeID struct{}

// ID ...
func (DefaultNodeID) ID(node *corev3.Node) string {
	return "default"
}
