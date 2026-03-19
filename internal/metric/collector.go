package metric

// Collector is a small abstraction for any metric source (CPU, memory, disk, etc).
type Collector interface {
	// Name is a stable key used in logs and alert state maps.
	Name() string
	// Init prepares internal state before first Sample call.
	Init() error
	// Sample returns current metric usage in percent (0-100).
	Sample() (float64, error)
}

const (
	NameCPU    = "cpu"
	NameMemory = "memory"
)
