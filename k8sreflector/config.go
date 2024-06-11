package k8sreflector

import (
	"time"
)

// ReflectorConfig ... reflector configuration
type ReflectorConfig struct {
	ResyncPeriod time.Duration
}

func (r ReflectorConfig) defaultConfigure() ReflectorConfig {
	if r.ResyncPeriod == 0 {
		r.ResyncPeriod = 5 * time.Minute
	}
	return r
}
