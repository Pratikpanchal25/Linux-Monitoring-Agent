package metric

import "linux-monitoring-agent/internal/cpu"

// CPUCollector samples CPU usage using /proc/stat deltas.
type CPUCollector struct {
	prev cpu.Snapshot
}

func NewCPUCollector() *CPUCollector {
	return &CPUCollector{}
}

func (c *CPUCollector) Name() string {
	return NameCPU
}

// Init captures the first baseline snapshot.
func (c *CPUCollector) Init() error {
	snapshot, err := cpu.ReadSnapshot()
	if err != nil {
		return err
	}
	c.prev = snapshot
	return nil
}

// Sample reads current CPU counters and calculates usage from deltas.
func (c *CPUCollector) Sample() (float64, error) {
	current, err := cpu.ReadSnapshot()
	if err != nil {
		return 0, err
	}

	usage, err := cpu.UsagePercent(c.prev, current)
	// Always move baseline forward to avoid stale deltas after transient errors.
	c.prev = current
	if err != nil {
		return 0, err
	}

	return usage, nil
}
