# Exercise 01 — container stats

Compute CPU% and memory% from raw Docker stats payloads. The streaming part —
`cli.ContainerStats(ctx, id, true)` returning an `io.ReadCloser` of
newline-delimited JSON — is trivial; the math is where it gets interesting.

## What you implement

```go
type Snapshot struct {
    Read              time.Time
    CPUTotalNanos     uint64
    SystemCPUNanos    uint64
    OnlineCPUs        uint32
    MemoryUsageBytes  uint64
    MemoryLimitBytes  uint64
}

func CPUPercent(curr, prev Snapshot) float64
func MemoryPercent(s Snapshot) float64
```

## The CPU formula

The daemon gives you **cumulative** counters. You compute deltas across two
samples:

```
cpu_delta    = curr.CPUTotalNanos   - prev.CPUTotalNanos      // ns of container CPU
system_delta = curr.SystemCPUNanos  - prev.SystemCPUNanos     // ns of total host CPU
pct          = (cpu_delta / system_delta) * onlineCPUs * 100
```

`onlineCPUs` is the multiplier because a container can use ALL cores; a
single-core formula caps at 100% even when the container is pegging 4 cores.
This is the same formula `docker stats` itself uses (see the daemon source).

## Contract pinned by tests

- 4 cores, 200ms CPU used over 1s wall → **20%**.
- Identical curr/prev (no time passed) → **0**, not NaN.
- Counter reset (curr < prev — container restarted) → **0**, not negative.
- First sample (prev is zero) → must not return NaN.
- Memory% with a limit set → percentage.
- Memory% with no limit (`MemoryLimitBytes == 0`) → **0** (don't divide by zero).

## Optional: wire it to a real daemon

Once you have the math right, the wiring is ~15 lines:

```go
rc, _ := cli.ContainerStats(ctx, id, true /* stream */)
dec := json.NewDecoder(rc)
var curr container.StatsResponse
var prev Snapshot
for dec.More() {
    _ = dec.Decode(&curr)
    s := toSnapshot(curr)
    cpu := CPUPercent(s, prev)
    fmt.Printf("CPU=%.1f%%  MEM=%.1f%%\n", cpu, MemoryPercent(s))
    prev = s
}
```

The mapping from the SDK's `container.StatsResponse` to our `Snapshot` is
straightforward — `CPUTotalNanos = stats.CPUStats.CPUUsage.TotalUsage`,
`SystemCPUNanos = stats.CPUStats.SystemUsage`, etc. Wiring this is up to you.

## Run the failing suite

```bash
go test -tags=exercise ./09-docker/exercises/01-container-stats/
```
