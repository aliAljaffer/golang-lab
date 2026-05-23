package main

import (
	"context"
	"errors"
	"fmt"
)

// Opts is the high-level request to Run: what to deploy and how.
type Opts struct {
	Owner         string
	Repo          string
	Tag           string
	HealthPath    string // appended to http://localhost:<HostPort>
	HostPort      int
	ContainerPort int
	Env           []string
	KeepContainer bool // if true, do NOT Remove on health failure
	DryRun        bool // if true, return a Report without calling Docker
}

// Report is what Run returns to its caller (the cobra cmd or a test).
//
// Fields are filled in incrementally as the pipeline progresses; on a
// mid-pipeline error the earlier fields are still meaningful (e.g. you
// know which stage failed because Healthy is false but ContainerID is set).
type Report struct {
	ArtifactURL string
	ImageID     string
	ContainerID string
	Healthy     bool
	Removed     bool // true iff this Run actually called Runner.Remove
}

// Run orchestrates the deploy pipeline:
//
//	1. fetcher.Fetch(owner, repo, tag) -> artifactURL
//	2. downloader.Download(artifactURL) -> tar bytes
//	3. builder.Build(tar, DockerSafeTag(owner, repo, tag)) -> imageID
//	4. runner.Run(RunOpts{ImageID, ports, env, RemoveOnExit: !KeepContainer}) -> containerID
//	5. health.Probe("http://localhost:<HostPort><HealthPath>") -> nil on healthy
//
// Failure handling:
//   - Any stage's error fails Run; later stages do not run.
//   - If health.Probe fails AND opts.KeepContainer == false, Run calls
//     runner.Remove(containerID) on a best-effort basis (its error is
//     written to the returned error message but does not replace the
//     original health failure).
//   - The Report is always returned (even on error), populated with
//     whatever stages did complete. This lets callers print a useful
//     status even when the deploy didn't go green.
//
// DryRun: if opts.DryRun, fetcher is still called (so a typo'd tag fails
// fast) but Download/Build/Run/Probe are skipped. Report.Healthy stays
// false; Report.ArtifactURL is populated.
func Run(
	ctx context.Context,
	fetcher ReleaseFetcher,
	downloader Downloader,
	builder Builder,
	runner Runner,
	health HealthChecker,
	opts Opts,
) (Report, error) {
	// TODO: walk the five-stage pipeline in the doc above. The test file
	//   verifies each branch — read it before you start writing.
	//
	//   Contract details that the tests pin:
	//     - the Report is returned on error too; populate it as you go so a
	//       caller can see how far the pipeline got. ImageID set + Healthy=false
	//       is what "build OK, container started, health failed" looks like.
	//     - errors at each stage should be wrapped with a stage tag
	//       (`fmt.Errorf("fetch: %w", err)`, etc.) so the test can assert which
	//       stage failed via errors.Is/As without string-matching.
	//     - DryRun short-circuits AFTER Fetch — a typo'd tag still fails fast.
	//     - on a health failure with KeepContainer=false, best-effort Remove the
	//       container and flip Report.Removed when it works. The health error is
	//       still what gets returned — Remove failing doesn't replace it.
	_ = fmt.Sprintf
	return Report{}, errors.New("Run not implemented")
}
