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
	// TODO: var report Report.
	//
	// TODO: url, err := fetcher.Fetch(ctx, opts.Owner, opts.Repo, opts.Tag).
	// TODO: if err != nil { return report, fmt.Errorf("fetch: %w", err) }.
	// TODO: report.ArtifactURL = url.
	//
	// TODO: if opts.DryRun { return report, nil }.
	//
	// TODO: body, err := downloader.Download(ctx, url).
	// TODO: if err != nil { return report, fmt.Errorf("download: %w", err) }.
	// TODO: defer body.Close().
	// TODO: tar, err := io.ReadAll(body).
	// TODO: if err != nil { return report, fmt.Errorf("read download body: %w", err) }.
	//
	// TODO: imageTag := DockerSafeTag(opts.Owner, opts.Repo, opts.Tag).
	// TODO: imageID, err := builder.Build(ctx, tar, imageTag).
	// TODO: if err != nil { return report, fmt.Errorf("build: %w", err) }.
	// TODO: report.ImageID = imageID.
	//
	// TODO: ctrID, err := runner.Run(ctx, RunOpts{
	// TODO:     ImageID: imageID, HostPort: opts.HostPort, ContainerPort: opts.ContainerPort,
	// TODO:     Env: opts.Env, RemoveOnExit: !opts.KeepContainer,
	// TODO: }).
	// TODO: if err != nil { return report, fmt.Errorf("run: %w", err) }.
	// TODO: report.ContainerID = ctrID.
	//
	// TODO: probeURL := fmt.Sprintf("http://localhost:%d%s", opts.HostPort, opts.HealthPath).
	// TODO: if err := health.Probe(ctx, probeURL); err != nil {
	// TODO:     if !opts.KeepContainer {
	// TODO:         if rmErr := runner.Remove(ctx, ctrID); rmErr == nil { report.Removed = true }
	// TODO:     }
	// TODO:     return report, fmt.Errorf("health: %w", err).
	// TODO: }
	//
	// TODO: report.Healthy = true.
	// TODO: return report, nil.
	_ = fmt.Sprintf // keep fmt live for when you add the wrapping
	return Report{}, errors.New("Run not implemented")
}
