package cpu

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const procStatPath = "/proc/stat"

// Snapshot keeps cumulative CPU counters from /proc/stat.
type Snapshot struct {
	Total uint64
	Idle  uint64
}

// ReadSnapshot reads the first "cpu" line from /proc/stat.
func ReadSnapshot() (Snapshot, error) {
	file, err := os.Open(procStatPath)
	if err != nil {
		return Snapshot{}, fmt.Errorf("open %s: %w", procStatPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return Snapshot{}, fmt.Errorf("scan %s: %w", procStatPath, err)
		}
		return Snapshot{}, fmt.Errorf("%s is empty", procStatPath)
	}

	line := scanner.Text()
	if !strings.HasPrefix(line, "cpu ") {
		return Snapshot{}, fmt.Errorf("unexpected first line in %s: %q", procStatPath, line)
	}

	fields := strings.Fields(line)
	if len(fields) < 5 {
		return Snapshot{}, fmt.Errorf("invalid cpu fields in %s", procStatPath)
	}

	var total uint64
	for _, raw := range fields[1:] {
		value, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return Snapshot{}, fmt.Errorf("parse cpu value %q: %w", raw, err)
		}
		total += value
	}

	idle, err := strconv.ParseUint(fields[4], 10, 64)
	if err != nil {
		return Snapshot{}, fmt.Errorf("parse idle value %q: %w", fields[4], err)
	}

	var iowait uint64
	if len(fields) > 5 {
		iowait, err = strconv.ParseUint(fields[5], 10, 64)
		if err != nil {
			return Snapshot{}, fmt.Errorf("parse iowait value %q: %w", fields[5], err)
		}
	}

	return Snapshot{Total: total, Idle: idle + iowait}, nil
}

// UsagePercent calculates CPU usage from two snapshots using deltas.
func UsagePercent(previous, current Snapshot) (float64, error) {
	if current.Total < previous.Total || current.Idle < previous.Idle {
		return 0, fmt.Errorf("cpu counters moved backwards")
	}

	deltaTotal := current.Total - previous.Total
	deltaIdle := current.Idle - previous.Idle
	if deltaTotal == 0 {
		return 0, fmt.Errorf("delta_total is zero")
	}
	if deltaIdle > deltaTotal {
		return 0, fmt.Errorf("delta_idle is greater than delta_total")
	}

	usage := (float64(deltaTotal-deltaIdle) / float64(deltaTotal)) * 100
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}

	return usage, nil
}
