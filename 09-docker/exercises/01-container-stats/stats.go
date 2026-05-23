// Package stats computes CPU% and memory% from Docker stats payloads.
//
// Exercise surface:
//
//	type Snapshot struct { ... raw daemon fields ... }
//	func CPUPercent(curr, prev Snapshot) float64
//	func MemoryPercent(s Snapshot) float64
//
// Why these two functions instead of "wrap ContainerStats"?
// Streaming stats from a real daemon is the easy part — `cli.ContainerStats(ctx, id, true)`
// returns an io.ReadCloser of newline-delimited JSON. Decoding it is trivial.
// The interesting work is the math: the daemon gives you raw cumulative
// counters, and you have to compute *deltas* yourself. That's the part where
// people get it wrong; that's what these tests pin.
package stats

import (
	"errors"
	"time"
)

// Snapshot is the slice of `container.StatsResponse` we need. Field names
// match the daemon JSON exactly. (The full type has dozens more fields —
// network, blkio, throttling — but these four are enough to compute the two
// percentages that 99% of consumers want.)
type Snapshot struct {
	Read              time.Time // when the daemon sampled this
	CPUTotalNanos     uint64    // container's cumulative CPU usage
	SystemCPUNanos    uint64    // host's cumulative CPU usage
	OnlineCPUs        uint32    // number of online CPUs on the host
	MemoryUsageBytes  uint64    // RSS used by the container
	MemoryLimitBytes  uint64    // cgroup hard limit (0 means "no limit")
}

// CPUPercent computes the container's CPU usage between two snapshots,
// as a percentage of all-cores capacity.
//
// Formula (the one Docker itself uses; see `docker stats`):
//
//	cpu_delta    = curr.CPUTotalNanos   - prev.CPUTotalNanos
//	system_delta = curr.SystemCPUNanos  - prev.SystemCPUNanos
//	pct          = (cpu_delta / system_delta) * onlineCPUs * 100.0
//
// Edge cases the tests pin:
//   - prev with zero counters (first sample) → return 0, NOT NaN.
//   - system_delta == 0 → return 0 (no time has passed; division by zero).
//   - curr.CPUTotalNanos < prev → counter rolled (container restarted);
//     return 0. Don't return a negative percentage.
func CPUPercent(curr, prev Snapshot) float64 {
	// TODO: implement the formula above. The interesting bit is the three
	//   edge cases that should return 0: the first sample (prev counters
	//   are zero, so cpuDelta is "the whole thing"), zero system delta
	//   (would divide by zero), and a wrapped/restarted counter (would
	//   underflow uint64 subtraction into a huge positive number).
	_ = errors.New
	return 0
}

// MemoryPercent computes the container's memory usage as a percentage of its
// memory limit. Returns 0 if no limit is set (limit == 0).
//
// Note: a container with no memory limit is common (it inherits the host's
// memory). Reporting "97% of unlimited" makes no sense; return 0 instead.
func MemoryPercent(s Snapshot) float64 {
	// TODO: percentage of MemoryUsageBytes / MemoryLimitBytes — but only
	//   when the limit is set. Containers without a cgroup limit return 0
	//   (97% of "unlimited" is meaningless).
	return 0
}
