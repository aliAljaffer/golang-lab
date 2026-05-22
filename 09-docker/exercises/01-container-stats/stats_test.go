//go:build exercise

package stats

import (
	"math"
	"testing"
	"time"
)

func TestCPUPercent_TypicalCase(t *testing.T) {
	// curr CPU = 200ms used in 1s of wall time, on 4 cores.
	// (1s wall on 4 cores = 4s of system CPU; so 200ms/4s = 5%.)
	prev := Snapshot{Read: time.Unix(0, 0), CPUTotalNanos: 0, SystemCPUNanos: 0, OnlineCPUs: 4}
	curr := Snapshot{Read: time.Unix(1, 0), CPUTotalNanos: 200_000_000, SystemCPUNanos: 4_000_000_000, OnlineCPUs: 4}

	got := CPUPercent(curr, prev)
	want := 20.0 // (0.2 / 4.0) * 4 * 100 == 20.0
	if math.Abs(got-want) > 0.001 {
		t.Errorf("CPUPercent = %f, want %f", got, want)
	}
}

func TestCPUPercent_NoTimePassed(t *testing.T) {
	s := Snapshot{CPUTotalNanos: 100, SystemCPUNanos: 1_000, OnlineCPUs: 2}
	got := CPUPercent(s, s) // identical: deltas == 0
	if got != 0 {
		t.Errorf("CPUPercent(same, same) = %f, want 0", got)
	}
}

func TestCPUPercent_CounterReset(t *testing.T) {
	prev := Snapshot{CPUTotalNanos: 999_999_999, SystemCPUNanos: 1_000_000_000, OnlineCPUs: 2}
	curr := Snapshot{CPUTotalNanos: 1_000_000, SystemCPUNanos: 2_000_000_000, OnlineCPUs: 2}

	got := CPUPercent(curr, prev)
	if got != 0 {
		t.Errorf("CPUPercent on reset = %f, want 0 (never negative)", got)
	}
}

func TestCPUPercent_FirstSample(t *testing.T) {
	// First sample: prev is the zero Snapshot. Expect 0, not NaN.
	curr := Snapshot{CPUTotalNanos: 100, SystemCPUNanos: 1_000, OnlineCPUs: 2}
	got := CPUPercent(curr, Snapshot{})
	if math.IsNaN(got) {
		t.Fatal("CPUPercent = NaN, want a real number (treat first sample as 0%)")
	}
	// systemDelta would be 1_000; cpuDelta would be 100; (100/1000)*2*100 = 20.
	// But the spec is: zero-counter prev is treated as "no previous data" -> 0.
	// We accept EITHER convention here — most tools return 0 for the first sample.
	// What we MUST NOT do is return NaN.
}

func TestMemoryPercent_HasLimit(t *testing.T) {
	s := Snapshot{MemoryUsageBytes: 256 << 20, MemoryLimitBytes: 1 << 30} // 256 MiB / 1 GiB
	got := MemoryPercent(s)
	want := 25.0
	if math.Abs(got-want) > 0.001 {
		t.Errorf("MemoryPercent = %f, want %f", got, want)
	}
}

func TestMemoryPercent_NoLimit(t *testing.T) {
	s := Snapshot{MemoryUsageBytes: 256 << 20, MemoryLimitBytes: 0}
	got := MemoryPercent(s)
	if got != 0 {
		t.Errorf("MemoryPercent on unlimited container = %f, want 0", got)
	}
}
