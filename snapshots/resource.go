package snapshots

import (
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	runtimev3 "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

func resourceType(res types.Resource) resourcev3.Type {
	switch res.(type) {
	case *listenerv3.Listener:
		return resourcev3.ListenerType
	case *routev3.RouteConfiguration:
		return resourcev3.RouteType
	case *clusterv3.Cluster:
		return resourcev3.ClusterType
	case *endpointv3.ClusterLoadAssignment:
		return resourcev3.EndpointType
	case *routev3.ScopedRouteConfiguration:
		return resourcev3.ScopedRouteType
	case *authv3.Secret:
		return resourcev3.SecretType
	case *runtimev3.Runtime:
		return resourcev3.RuntimeType
	case *corev3.TypedExtensionConfig:
		return resourcev3.ExtensionConfigType
	default:
		return ""
	}
}

func resourcesToMap(resources []types.Resource) map[string][]types.Resource {
	out := map[string][]types.Resource{}
	for _, res := range resources {
		rt := resourceType(res)
		if _, ok := out[rt]; !ok {
			out[rt] = []types.Resource{res}
		} else {
			out[rt] = append(out[rt], res)
		}
	}
	return out
}
