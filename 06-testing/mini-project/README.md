# Mini-project — `logstats`

A small log aggregator that intentionally has surface area for every testing pattern from examples 01-08.

## Spec

```
logstats <source>            # source = file path OR http(s):// URL
logstats --format json <source>
```

Reads lines like `[INFO] request handled in 12ms`, counts per level, reports a throughput rate.

Sample run:

```
$ logstats ./testdata/lines.log
total: 10
  DEBUG 1
  INFO  5
  WARN  2
  ERROR 2
rate:  500.0 events/s
```

## How it's split for testability

| Piece | Type | Used in tests for |
|---|---|---|
| `Parse(line)` | pure | table tests (02), subtests (03), benchmark (07), fuzz (08) |
| `FormatRate(n, d)` | pure | subtests (03) |
| `Aggregator.Add/Snapshot` | stateful | helper-with-`t.Helper()` |
| `Source` interface | seam | hand-rolled fake (04) |
| `FileSource.Fetch` | I/O | fixtures from `testdata/` (06) |
| `HTTPSource.Fetch` | I/O | `httptest.NewServer` (05) |
| `Summarize(ctx, src)` | composition | end-to-end through all of the above |
| `TestMain` | suite-level | demonstrates package-level setup/teardown |

The test file (`main_test.go`) is annotated with which example concept each test or block demonstrates — read it after working through examples 01-08 to see the patterns combined.

## Run the tests

```
# all tests (some fail until you implement the TODOs)
go test -tags=exercise ./06-testing/mini-project/...

# verbose, see TestMain output + every subtest name
go test -tags=exercise -v ./06-testing/mini-project/...

# benchmark
go test -tags=exercise -bench=BenchmarkParse_1k -benchmem ./06-testing/mini-project/...

# fuzz interactively (writes failing inputs to testdata/fuzz/)
go test -tags=exercise -fuzz=FuzzParse -fuzztime=10s ./06-testing/mini-project/...

# coverage report
go test -tags=exercise -coverprofile=/tmp/c.out ./06-testing/mini-project/...
go tool cover -html=/tmp/c.out
```

## Note on PLAN.md

`PLAN.md` originally framed the 06-testing mini-project as "add tests retroactively to `dirsize` and `gh-repo-stats`." Those tests already exist (added during sections 01 and 03 scaffolding) so this mini-project instead consolidates every pattern into one place. Use it as the kitchen-sink reference.

## Stretch ideas

- Add a benchmark that compares `Summarize` against a 1 MB file vs an 11 MB file. Is it linear?
- Replace the `bufio.Scanner` with a `bufio.Reader` and benchmark the difference.
- Track per-level rate (events/sec for ERROR only) in the snapshot.
- Add a `--tail` flag that streams stats every N seconds (similar to `tail -f` from `02-files-and-os`).
- Add a `t.Parallel()` to the table subtests in `TestParse` and confirm output order scrambles.
