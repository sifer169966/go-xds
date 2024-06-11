package monitor

import (
	"encoding/json"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/encoding/protojson"
)

type cacheMarshaler struct {
	cachev3.Cache
}

func (c cacheMarshaler) MarshalJSON() ([]byte, error) {
	snapshotCache, ok := c.Cache.(cachev3.SnapshotCache)
	if !ok {
		return nil, nil
	}
	out := map[string]map[resourcev3.Type]interface{}{}
	nodeIDs := snapshotCache.GetStatusKeys()
	if len(nodeIDs) == 0 {
		nodeIDs = append(nodeIDs, "default")
	}
	for _, nodeID := range nodeIDs {
		snapshot, err := snapshotCache.GetSnapshot(nodeID)
		if err != nil {
			return nil, err
		}
		nodeResources := map[resourcev3.Type]interface{}{}
		for i := types.ResponseType(0); i < types.UnknownType; i++ {
			typeURL, _ := cachev3.GetResponseTypeURL(i)
			resources := snapshot.GetResources(typeURL)
			if len(resources) == 0 {
				continue
			}
			version := snapshot.GetVersion(typeURL)
			nodeResources[typeURL] = nodeResourcesMarshaler{
				version:   version,
				resources: resources,
			}
		}
		out[nodeID] = nodeResources
	}
	return json.Marshal(out)
}

type nodeResourcesMarshaler struct {
	version   string
	resources map[string]types.Resource
}

func (r nodeResourcesMarshaler) MarshalJSON() ([]byte, error) {
	type nodeResource struct {
		Version   string                       `json:"version"`
		Resources map[string]resourceMarshaler `json:"resources"`
	}
	out := nodeResource{
		Version:   r.version,
		Resources: map[string]resourceMarshaler{},
	}

	for k, v := range r.resources {
		out.Resources[k] = resourceMarshaler{Resource: v}
	}

	return json.Marshal(out)
}

type resourceMarshaler struct {
	types.Resource
}

func (r resourceMarshaler) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(r.Resource)
}
