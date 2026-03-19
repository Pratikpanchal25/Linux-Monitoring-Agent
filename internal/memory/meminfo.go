package memory

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const meminfoPath = "/proc/meminfo"

// UsagePercent returns RAM usage as: (MemTotal - MemAvailable) / MemTotal * 100.
// If MemAvailable is missing, a fallback approximation is used.
func UsagePercent() (float64, error) {
	file, err := os.Open(meminfoPath)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", meminfoPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var memTotal uint64
	var memAvailable uint64
	var memFree uint64
	var buffers uint64
	var cached uint64
	var sReclaimable uint64
	var shmem uint64

	hasTotal := false
	hasAvailable := false
	hasFallbackParts := false

	for scanner.Scan() {
		line := scanner.Text()
		key, rest, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}

		value, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse %s value %q: %w", key, fields[0], err)
		}

		switch key {
		case "MemTotal":
			memTotal = value
			hasTotal = true
		case "MemAvailable":
			memAvailable = value
			hasAvailable = true
		case "MemFree":
			memFree = value
			hasFallbackParts = true
		case "Buffers":
			buffers = value
			hasFallbackParts = true
		case "Cached":
			cached = value
			hasFallbackParts = true
		case "SReclaimable":
			sReclaimable = value
			hasFallbackParts = true
		case "Shmem":
			shmem = value
			hasFallbackParts = true
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("scan %s: %w", meminfoPath, err)
	}
	if !hasTotal || memTotal == 0 {
		return 0, fmt.Errorf("MemTotal missing or zero in %s", meminfoPath)
	}

	available := memAvailable
	if !hasAvailable {
		if !hasFallbackParts {
			return 0, fmt.Errorf("MemAvailable missing and fallback fields unavailable in %s", meminfoPath)
		}
		// Fallback approximation for older kernels lacking MemAvailable.
		available = memFree + buffers + cached + sReclaimable
		if available > shmem {
			available -= shmem
		} else {
			available = 0
		}
	}

	if available > memTotal {
		available = memTotal
	}

	used := memTotal - available
	usage := (float64(used) / float64(memTotal)) * 100
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}

	return usage, nil
}
