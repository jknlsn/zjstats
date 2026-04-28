package collector

import (
	"github.com/jknlsn/zjstat/internal/metrics"
)

// Collector gathers system metrics.
type Collector interface {
	Collect() (*metrics.Snapshot, error)
}

// New returns a platform-appropriate Collector.
func New() Collector {
	return newDarwinCollector()
}
