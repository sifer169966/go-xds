package snapshots

import (
	"context"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"k8s.io/klog/v2"
)

const (
	resourceKindLDS   = "LDS"
	resourceKindRDS   = "RDS"
	resourceKindCDS   = "CDS"
	resourceKindEDS   = "EDS"
	resourceKindMixed = "LDS/RDS/CDS"
)

type SnapshotSetter interface {
	Set(ctx context.Context, version string, src []types.Resource)
}

// Snapshot ...
type Snapshot struct {
	muxCache           cachev3.MuxCache
	mixedSnapshotCache cachev3.SnapshotCache
	edsSnapshotCache   cachev3.SnapshotCache
}

func getResourceKeyName(typeURL string) string {
	switch typeURL {
	case resourcev3.ListenerType, resourcev3.RouteType, resourcev3.ClusterType:
		return resourceKindMixed
	case resourcev3.EndpointType:
		return resourceKindEDS
	default:
		return ""
	}
}

// New ...
// create a new instance of snapshot to capture and hold the discovery information at a point of time
func New() *Snapshot {
	mixedSnapshotCache := cachev3.NewSnapshotCache(false, DefaultNodeID{}, nil)
	edsSnapshotCache := cachev3.NewSnapshotCache(false, DefaultNodeID{}, nil)
	muxCache := cachev3.MuxCache{
		Classify: func(r *cachev3.Request) string {
			return getResourceKeyName(r.TypeUrl)
		},
		ClassifyDelta: func(r *cachev3.DeltaRequest) string {
			return getResourceKeyName(r.TypeUrl)
		},
		Caches: map[string]cachev3.Cache{
			resourceKindMixed: mixedSnapshotCache,
			resourceKindEDS:   edsSnapshotCache,
		},
	}
	return &Snapshot{
		muxCache:           muxCache,
		mixedSnapshotCache: mixedSnapshotCache,
		edsSnapshotCache:   edsSnapshotCache,
	}
}

func (s *Snapshot) MuxCache() *cachev3.MuxCache {
	return &s.muxCache
}

// Set ...
// set the mixed snapshot(multiplex of LDS, RDS, CDS) and a separate snapshot for eds
// if the src is EDS then set the EDS snapshot, otherwise, set the resource into mixed snapshot
func (s *Snapshot) Set(ctx context.Context, version string, src []types.Resource) {
	srcMap := resourcesToMap(src)
	snapshot, err := cachev3.NewSnapshot(version, srcMap)
	if err != nil {
		klog.Error("could not create a new snapshot", "version", version, "src", srcMap)
		return
	}
	//TODO: hasing resources to compare with the previous snapshot
	if _, ok := srcMap[resourcev3.EndpointType]; ok {
		s.setEDSSnapshotCache(ctx, snapshot)
		klog.Info("set eds snapshot to a new version", "version", version)
	} else {
		s.setMixedSnapshotCache(ctx, snapshot)
		klog.Info("set mixed snapshot to a new version", "version", version)
	}
}

func (s *Snapshot) setEDSSnapshotCache(ctx context.Context, snap *cachev3.Snapshot) {
	nodeIDs := s.edsSnapshotCache.GetStatusKeys()
	// predefined the node ID in case there is no request from the client yet to provide the information for our monitoring
	if len(nodeIDs) == 0 {
		nodeIDs = append(nodeIDs, DefaultNodeID{}.ID(nil))
	}
	for _, v := range nodeIDs {
		s.edsSnapshotCache.SetSnapshot(ctx, v, snap)
	}
}

func (s *Snapshot) setMixedSnapshotCache(ctx context.Context, snap *cachev3.Snapshot) {
	nodeIDs := s.mixedSnapshotCache.GetStatusKeys()
	// predefined the node ID in case there is no request from the client yet to provide the information for our monitoring
	if len(nodeIDs) == 0 {
		nodeIDs = append(nodeIDs, DefaultNodeID{}.ID(nil))
	}
	for _, v := range nodeIDs {
		s.mixedSnapshotCache.SetSnapshot(ctx, v, snap)
	}
}
