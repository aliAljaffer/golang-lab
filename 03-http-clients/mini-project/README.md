# `gh-repo-stats` — mini-project

Fetch GitHub repo metadata (stars, forks, last push) for a list of repos.
Cache responses (ETag) to avoid hammering the rate limit. Retry on 5xx / 429.
Output CSV.

## Spec

```bash
gh-repo-stats --repos golang/go,spf13/cobra            # writes to stdout
gh-repo-stats --repos golang/go --cache gh.json -o stats.csv
```

Output schema:

```csv
name,stars,forks,pushed_at
golang/go,120000,17000,2026-05-19T12:00:00Z
```

## How to implement

The scaffold splits the work into testable functions:

| Function                                  | Job                                                                         |
| ----------------------------------------- | --------------------------------------------------------------------------- |
| `fetchStats(client, baseURL, repo, etag)` | One repo. Send `If-None-Match`. Decode JSON or return `fresh=false` on 304. |
| `doWithRetry(client, req, maxAttempts)`   | Retry transport errors, 5xx, 429 with exponential backoff + jitter.         |
| `loadCache(path)` / `saveCache(path, c)`  | Persist `{repo: {etag, stats}}` as JSON between runs.                       |
| `writeCSV(w, rows)`                       | Header + rows. Used both for `--out file.csv` and `--out -`.                |
| `newRootCmd()`                            | cobra wiring + orchestration.                                               |

`baseURL` is a parameter (not a constant) so the tests can point the client at an `httptest.NewServer` and verify retry/ETag behaviour without touching the real GitHub API.

## Run the tests

```bash
go test -tags=exercise ./03-http-clients/mini-project/...
```

All tests fail until you implement the TODOs.

## Stretch ideas

- Honor `Retry-After` on 429 responses (parse the header — it can be an integer or an HTTP-date).
- Fetch repos concurrently with a bounded worker pool (`go func` + a `sem chan struct{}` of size N).
- Add a `--progress` flag that prints a one-line tally as repos complete.
- Wire a real token from `GH_TOKEN` to get the 5,000/hr rate limit instead of 60/hr.
