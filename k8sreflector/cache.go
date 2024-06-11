package k8sreflector

import "sync"

type localCache struct {
	// lastResourcesHash is the sum of resource hashed
	// it is thread safe, but not synchronized with the underlying store
	lastResourcesHash uint64
	// lastResourceHashMutex guards read/write access to lastResourceHashMutex
	lastResourceHashMutex sync.Mutex
}
