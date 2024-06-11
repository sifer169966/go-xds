package k8sreflector

import (
	"cmp"
	"slices"

	"github.com/cespare/xxhash/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"google.golang.org/protobuf/proto"
)

// use as a separator to separate each part of bytes. `\xff` equal to Ascii 255
var resourceSeparator = []byte{'\xff'}

// resourceHash hashing of the resources to use as a comparison against the latest resources
func resourceHash(resources []types.Resource) (uint64, error) {
	slices.SortStableFunc(resources, func(a, b types.Resource) int {
		return cmp.Compare(cachev3.GetResourceName(a), cachev3.GetResourceName(b))
	})
	b := []byte{}
	var err error
	h := xxhash.New()
	for _, resource := range resources {
		b, err = proto.MarshalOptions{
			Deterministic: true,
		}.MarshalAppend(b, resource)
		if err != nil {
			return 0, err
		}
		h.Write(b)
		h.Write(resourceSeparator)
		b = b[:0]
	}
	return h.Sum64(), nil
}
