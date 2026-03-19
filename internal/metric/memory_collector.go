package metric

import "linux-monitoring-agent/internal/memory"

// MemoryCollector samples RAM usage using /proc/meminfo.
type MemoryCollector struct{}

func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{}
}

func (m *MemoryCollector) Name() string {
	return NameMemory
}

func (m *MemoryCollector) Init() error {
	// No baseline needed for memory usage.
	return nil
}

func (m *MemoryCollector) Sample() (float64, error) {
	return memory.UsagePercent()
}
