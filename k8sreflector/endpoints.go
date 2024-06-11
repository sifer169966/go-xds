package k8sreflector

import (
	"context"
	"fmt"
	"sort"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/sifer169966/go-xds/snapshots"
	"google.golang.org/protobuf/types/known/wrapperspb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	k8scache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type EndpointReflector struct {
	api        kubernetes.Interface
	snap       snapshots.SnapshotSetter
	refl       *k8scache.Reflector
	localCache localCache
	cfg        ReflectorConfig
}

// NewEndpointReflector ... create a new instance of *EndpointReflector
func NewEndpointReflector(c kubernetes.Interface, s snapshots.SnapshotSetter, cfg ReflectorConfig) *EndpointReflector {
	return &EndpointReflector{
		api:  c,
		snap: s,
		cfg:  cfg.defaultConfigure(),
	}
}

// Watch ... run the reflector to watching against k8s API to get the information about endpoint resources
func (r *EndpointReflector) Watch(ctx context.Context) error {
	store := k8scache.NewUndeltaStore(r.endpointsPushFunc(ctx), k8scache.DeletionHandlingMetaNamespaceKeyFunc)
	r.refl = k8scache.NewReflector(&k8scache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return r.api.CoreV1().Endpoints("").List(ctx, opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return r.api.CoreV1().Endpoints("").Watch(ctx, opts)
		},
	}, &corev1.Endpoints{}, store, r.cfg.ResyncPeriod)
	klog.Info("starting endpoints reflector")
	r.refl.Run(ctx.Done())
	klog.Warning("endpoints reflector has been stopped")
	return nil
}

func (r *EndpointReflector) endpointsPushFunc(ctx context.Context) func(v []interface{}) {
	return func(v []interface{}) {
		latestVersion := r.refl.LastSyncResourceVersion()
		endpoints := sliceToEndpoints(v)
		resources := endpointsToResources(endpoints)
		resourcesHashed, err := resourceHash(resources)
		if err == nil {
			r.localCache.lastResourceHashMutex.Lock()
			defer r.localCache.lastResourceHashMutex.Unlock()
			if resourcesHashed == r.localCache.lastResourcesHash {
				klog.Info("endpoint resources hashed equal with the previous one, no need to update")
				return
			}
			r.localCache.lastResourcesHash = resourcesHashed
		} else {
			klog.Error("endpoint resource hash failed", "err", err)
		}
		r.snap.Set(ctx, latestVersion, resources)
	}
}

// endpointsToResources ...
// creating eds resources from k8s endpoints
func endpointsToResources(eps []*corev1.Endpoints) []types.Resource {
	var out []types.Resource
	for _, ep := range eps {
		for _, subset := range ep.Subsets {
			for _, port := range subset.Ports {
				var clusterName string
				if port.Name == "" {
					clusterName = fmt.Sprintf("%s.%s:%d", ep.Name, ep.Namespace, port.Port)
				} else {
					clusterName = fmt.Sprintf("%s.%s:%s", ep.Name, ep.Namespace, port.Name)
				}
				cla := &endpointv3.ClusterLoadAssignment{
					ClusterName: clusterName,
					Endpoints: []*endpointv3.LocalityLbEndpoints{
						{
							LoadBalancingWeight: wrapperspb.UInt32(1),
							Locality:            &corev3.Locality{},
							LbEndpoints:         []*endpointv3.LbEndpoint{},
						},
					},
				}
				sort.SliceStable(subset.Addresses, func(i, j int) bool {
					left := subset.Addresses[i].IP
					right := subset.Addresses[j].IP
					return left < right
				})
				for _, addr := range subset.Addresses {
					hostname := addr.Hostname
					if hostname == "" && addr.TargetRef != nil {
						hostname = fmt.Sprintf("%s.%s", addr.TargetRef.Name, addr.TargetRef.Namespace)
					}
					if hostname == "" && addr.NodeName != nil {
						hostname = *addr.NodeName
					}
					cla.Endpoints[0].LbEndpoints = append(cla.Endpoints[0].LbEndpoints, &endpointv3.LbEndpoint{
						HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
							Endpoint: &endpointv3.Endpoint{
								Address: &corev3.Address{
									Address: &corev3.Address_SocketAddress{
										SocketAddress: &corev3.SocketAddress{
											Protocol: corev3.SocketAddress_TCP,
											Address:  addr.IP,
											PortSpecifier: &corev3.SocketAddress_PortValue{
												PortValue: uint32(port.Port),
											},
										},
									},
								},
								Hostname: hostname,
							},
						},
					})
				}
				out = append(out, cla)
			}
		}
	}
	return out
}

func sliceToEndpoints(endpoints []interface{}) []*corev1.Endpoints {
	out := make([]*corev1.Endpoints, len(endpoints))
	for i, ep := range endpoints {
		out[i] = ep.(*corev1.Endpoints)
	}
	return out
}
