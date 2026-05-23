//go:build exercise

package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ---- GHReleaseFetcher -----------------------------------------------------

// newFetcherServer returns an httptest.Server that serves /repos/:owner/:repo/releases/tags/:tag.
// The handler is supplied by the test; the server's URL is fed to GHReleaseFetcher.BaseURL.
func newFetcherServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestGHReleaseFetcher_NotFound(t *testing.T) {
	srv := newFetcherServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	f := &GHReleaseFetcher{HTTPClient: http.DefaultClient, BaseURL: srv.URL}

	_, err := f.Fetch(context.Background(), "acme", "thing", "v1")
	if !errors.Is(err, ErrReleaseNotFound) {
		t.Errorf("Fetch on 404 returned %v; want ErrReleaseNotFound", err)
	}
}

func TestGHReleaseFetcher_PrefersDockerfileAsset(t *testing.T) {
	const body = `{
		"tag_name": "v1.2.3",
		"assets": [
			{"name": "extra.tar.gz", "browser_download_url": "https://example.invalid/tar"},
			{"name": "Dockerfile",   "browser_download_url": "https://example.invalid/dockerfile"},
			{"name": "checksums.txt","browser_download_url": "https://example.invalid/checksums"}
		]
	}`
	srv := newFetcherServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, body)
	})
	f := &GHReleaseFetcher{HTTPClient: http.DefaultClient, BaseURL: srv.URL}

	url, err := f.Fetch(context.Background(), "acme", "thing", "v1.2.3")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if url != "https://example.invalid/dockerfile" {
		t.Errorf("picked %q; want the Dockerfile asset (\"https://example.invalid/dockerfile\")", url)
	}
}

func TestGHReleaseFetcher_FallsBackToTarball(t *testing.T) {
	const body = `{
		"tag_name": "v1.2.3",
		"assets": [
			{"name": "checksums.txt","browser_download_url": "https://example.invalid/checksums"},
			{"name": "source.tar.gz","browser_download_url": "https://example.invalid/source"}
		]
	}`
	srv := newFetcherServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, body)
	})
	f := &GHReleaseFetcher{HTTPClient: http.DefaultClient, BaseURL: srv.URL}

	url, err := f.Fetch(context.Background(), "acme", "thing", "v1.2.3")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if url != "https://example.invalid/source" {
		t.Errorf("picked %q; want the .tar.gz fallback (\"https://example.invalid/source\")", url)
	}
}

func TestGHReleaseFetcher_NoSuitableAsset(t *testing.T) {
	const body = `{
		"tag_name": "v1.2.3",
		"assets": [
			{"name": "checksums.txt","browser_download_url": "https://example.invalid/checksums"},
			{"name": "README.md",    "browser_download_url": "https://example.invalid/readme"}
		]
	}`
	srv := newFetcherServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, body)
	})
	f := &GHReleaseFetcher{HTTPClient: http.DefaultClient, BaseURL: srv.URL}

	_, err := f.Fetch(context.Background(), "acme", "thing", "v1.2.3")
	if !errors.Is(err, ErrNoSuitableAsset) {
		t.Errorf("Fetch with no Dockerfile/.tar.gz returned %v; want ErrNoSuitableAsset", err)
	}
}

func TestGHReleaseFetcher_SendsToken(t *testing.T) {
	var seenAuth string
	srv := newFetcherServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		_, _ = io.WriteString(w, `{"tag_name":"v1","assets":[{"name":"Dockerfile","browser_download_url":"x"}]}`)
	})
	f := &GHReleaseFetcher{HTTPClient: http.DefaultClient, BaseURL: srv.URL, Token: "secret-token"}

	if _, err := f.Fetch(context.Background(), "acme", "thing", "v1"); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if want := "Bearer secret-token"; seenAuth != want {
		t.Errorf("Authorization header = %q; want %q", seenAuth, want)
	}
}

// ---- DockerSafeTag --------------------------------------------------------

func TestDockerSafeTag(t *testing.T) {
	cases := []struct {
		name              string
		owner, repo, tag  string
		want              string
	}{
		{"plain", "acme", "thing", "v1.2.3", "acme-thing:v1.2.3"},
		{"uppercase owner", "Acme", "thing", "v1", "acme-thing:v1"},
		{"slash in repo", "acme", "my/repo", "v1", "acme-my-repo:v1"},
		{"uppercase tag is lowered", "acme", "thing", "V1.2.3", "acme-thing:v1.2.3"},
		{"slash in tag becomes dash", "acme", "thing", "feature/foo", "acme-thing:feature-foo"},
		{"leading dot in tag becomes underscore", "acme", "thing", ".hidden", "acme-thing:_hidden"},
		{"leading dash in tag becomes underscore", "acme", "thing", "-rc1", "acme-thing:_rc1"},
		{"odd chars get dashed", "acme", "thing", "v1+meta", "acme-thing:v1-meta"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DockerSafeTag(tc.owner, tc.repo, tc.tag); got != tc.want {
				t.Errorf("DockerSafeTag(%q, %q, %q) = %q; want %q", tc.owner, tc.repo, tc.tag, got, tc.want)
			}
		})
	}
}

// ---- DockerRunner (AutoRemove pass-through) ------------------------------

// fakeDockerSDK records what containerCreate received so the test can
// assert opts.RemoveOnExit was propagated to the SDK as autoRemove.
type fakeDockerSDK struct {
	mu sync.Mutex

	// captured args from containerCreate
	createImage         string
	createEnv           []string
	createHostPort      int
	createContainerPort int
	createAutoRemove    bool

	// programmed returns
	createID  string
	createErr error
	startErr  error
	removeErr error

	// call counts
	createCalls atomic.Int64
	startCalls  atomic.Int64
	removeCalls atomic.Int64
}

func (f *fakeDockerSDK) containerCreate(_ context.Context, image string, env []string, hostPort, containerPort int, autoRemove bool) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createCalls.Add(1)
	f.createImage = image
	f.createEnv = env
	f.createHostPort = hostPort
	f.createContainerPort = containerPort
	f.createAutoRemove = autoRemove
	if f.createID == "" {
		return "ctr-default", f.createErr
	}
	return f.createID, f.createErr
}

func (f *fakeDockerSDK) containerStart(_ context.Context, _ string) error {
	f.startCalls.Add(1)
	return f.startErr
}

func (f *fakeDockerSDK) containerRemove(_ context.Context, _ string) error {
	f.removeCalls.Add(1)
	return f.removeErr
}

func TestDockerRunner_PassesAutoRemove(t *testing.T) {
	sdk := &fakeDockerSDK{createID: "ctr-xyz"}
	r := &DockerRunner{Inner: sdk}

	_, err := r.Run(context.Background(), RunOpts{
		ImageID:       "img-1",
		HostPort:      8080,
		ContainerPort: 8080,
		Env:           []string{"FOO=bar"},
		RemoveOnExit:  true,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !sdk.createAutoRemove {
		t.Error("autoRemove arg to containerCreate = false; want true (RunOpts.RemoveOnExit was true)")
	}
	if sdk.createImage != "img-1" {
		t.Errorf("image arg = %q; want %q", sdk.createImage, "img-1")
	}
	if sdk.startCalls.Load() != 1 {
		t.Errorf("containerStart calls = %d; want 1", sdk.startCalls.Load())
	}
}

func TestDockerRunner_RemoveOrphansOnStartFailure(t *testing.T) {
	sdk := &fakeDockerSDK{createID: "ctr-xyz", startErr: errors.New("kaboom")}
	r := &DockerRunner{Inner: sdk}

	_, err := r.Run(context.Background(), RunOpts{ImageID: "img-1", RemoveOnExit: false})
	if err == nil {
		t.Fatal("Run returned nil error on a start failure")
	}
	if sdk.removeCalls.Load() != 1 {
		t.Errorf("containerRemove calls = %d; want 1 (orphan cleanup after start failure)", sdk.removeCalls.Load())
	}
}

// ---- HTTPHealthChecker ---------------------------------------------------

func fastSleep(ctx context.Context, _ time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func TestHTTPHealthChecker_SucceedsAfterNProbes(t *testing.T) {
	var calls atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	h := &HTTPHealthChecker{Client: http.DefaultClient, Interval: time.Millisecond, Sleep: fastSleep}
	if err := h.Probe(context.Background(), srv.URL); err != nil {
		t.Errorf("Probe = %v; want nil after 3rd attempt returned 200", err)
	}
	if calls.Load() != 3 {
		t.Errorf("probe attempts = %d; want 3", calls.Load())
	}
}

func TestHTTPHealthChecker_TimesOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	h := &HTTPHealthChecker{Client: http.DefaultClient, Interval: time.Millisecond, Sleep: fastSleep}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := h.Probe(ctx, srv.URL)
	if err == nil {
		t.Fatal("Probe returned nil against a never-2xx server with an expired ctx")
	}
}

func TestHTTPHealthChecker_PropagatesCtxCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(context.Background())
	// Sleep slice that ALWAYS returns ctx.Err() — simulates a cancelled
	// long sleep without actually waiting.
	cancellingSleep := func(_ context.Context, _ time.Duration) error {
		cancel()
		return ctx.Err()
	}
	h := &HTTPHealthChecker{Client: http.DefaultClient, Interval: time.Hour, Sleep: cancellingSleep}

	err := h.Probe(ctx, srv.URL)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Probe = %v; want context.Canceled propagation", err)
	}
}

// ---- end-to-end Run ------------------------------------------------------

type fakeFetcher struct {
	url   string
	err   error
	calls atomic.Int64
}

func (f *fakeFetcher) Fetch(_ context.Context, _, _, _ string) (string, error) {
	f.calls.Add(1)
	return f.url, f.err
}

type fakeDownloader struct {
	body  string
	err   error
	calls atomic.Int64
}

func (d *fakeDownloader) Download(_ context.Context, _ string) (io.ReadCloser, error) {
	d.calls.Add(1)
	if d.err != nil {
		return nil, d.err
	}
	return io.NopCloser(strings.NewReader(d.body)), nil
}

type fakeBuilder struct {
	imageID string
	err     error
	calls   atomic.Int64
	gotTag  string
}

func (b *fakeBuilder) Build(_ context.Context, _ []byte, tag string) (string, error) {
	b.calls.Add(1)
	b.gotTag = tag
	return b.imageID, b.err
}

type fakeRunner struct {
	containerID string
	runErr      error
	removeErr   error

	runCalls    atomic.Int64
	removeCalls atomic.Int64
	gotOpts     RunOpts
}

func (r *fakeRunner) Run(_ context.Context, opts RunOpts) (string, error) {
	r.runCalls.Add(1)
	r.gotOpts = opts
	return r.containerID, r.runErr
}

func (r *fakeRunner) Remove(_ context.Context, _ string) error {
	r.removeCalls.Add(1)
	return r.removeErr
}

type fakeHealth struct {
	err   error
	calls atomic.Int64
}

func (h *fakeHealth) Probe(_ context.Context, _ string) error {
	h.calls.Add(1)
	return h.err
}

func TestRun_HappyPath(t *testing.T) {
	f := &fakeFetcher{url: "https://example.invalid/x"}
	d := &fakeDownloader{body: "tarbytes"}
	b := &fakeBuilder{imageID: "img-1"}
	r := &fakeRunner{containerID: "ctr-1"}
	h := &fakeHealth{}

	report, err := Run(context.Background(), f, d, b, r, h, Opts{
		Owner: "acme", Repo: "thing", Tag: "v1",
		HostPort: 8080, ContainerPort: 8080, HealthPath: "/healthz",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !report.Healthy {
		t.Error("report.Healthy = false; want true on happy path")
	}
	if report.ImageID != "img-1" || report.ContainerID != "ctr-1" {
		t.Errorf("report = %+v; want ImageID=img-1, ContainerID=ctr-1", report)
	}
	if f.calls.Load() != 1 || d.calls.Load() != 1 || b.calls.Load() != 1 || r.runCalls.Load() != 1 || h.calls.Load() != 1 {
		t.Errorf("call counts = fetch:%d dl:%d build:%d run:%d health:%d; want 1 each",
			f.calls.Load(), d.calls.Load(), b.calls.Load(), r.runCalls.Load(), h.calls.Load())
	}
	if b.gotTag != "acme-thing:v1" {
		t.Errorf("Build received tag %q; want %q (DockerSafeTag)", b.gotTag, "acme-thing:v1")
	}
}

func TestRun_HealthFailureRemovesContainer(t *testing.T) {
	f := &fakeFetcher{url: "https://example.invalid/x"}
	d := &fakeDownloader{body: "tarbytes"}
	b := &fakeBuilder{imageID: "img-1"}
	r := &fakeRunner{containerID: "ctr-1"}
	h := &fakeHealth{err: errors.New("never 200")}

	report, err := Run(context.Background(), f, d, b, r, h, Opts{
		Owner: "acme", Repo: "thing", Tag: "v1",
		KeepContainer: false,
	})
	if err == nil {
		t.Fatal("Run returned nil on health failure")
	}
	if r.removeCalls.Load() != 1 {
		t.Errorf("Runner.Remove calls = %d; want 1 (cleanup on health fail with KeepContainer=false)", r.removeCalls.Load())
	}
	if !report.Removed {
		t.Error("report.Removed = false; want true after a successful cleanup")
	}
}

func TestRun_KeepContainerSkipsRemove(t *testing.T) {
	f := &fakeFetcher{url: "https://example.invalid/x"}
	d := &fakeDownloader{body: "tarbytes"}
	b := &fakeBuilder{imageID: "img-1"}
	r := &fakeRunner{containerID: "ctr-1"}
	h := &fakeHealth{err: errors.New("never 200")}

	_, err := Run(context.Background(), f, d, b, r, h, Opts{
		Owner: "acme", Repo: "thing", Tag: "v1",
		KeepContainer: true,
	})
	if err == nil {
		t.Fatal("Run returned nil on health failure")
	}
	if r.removeCalls.Load() != 0 {
		t.Errorf("Runner.Remove calls = %d; want 0 (KeepContainer=true)", r.removeCalls.Load())
	}
}

func TestRun_FetcherError_FailsFast(t *testing.T) {
	f := &fakeFetcher{err: ErrReleaseNotFound}
	d := &fakeDownloader{body: "tarbytes"}
	b := &fakeBuilder{imageID: "img-1"}
	r := &fakeRunner{containerID: "ctr-1"}
	h := &fakeHealth{}

	_, err := Run(context.Background(), f, d, b, r, h, Opts{Owner: "acme", Repo: "thing", Tag: "v1"})
	if !errors.Is(err, ErrReleaseNotFound) {
		t.Errorf("Run = %v; want a wrapped ErrReleaseNotFound", err)
	}
	if d.calls.Load() != 0 || b.calls.Load() != 0 || r.runCalls.Load() != 0 || h.calls.Load() != 0 {
		t.Errorf("downstream stages called after fetcher error: dl:%d build:%d run:%d health:%d",
			d.calls.Load(), b.calls.Load(), r.runCalls.Load(), h.calls.Load())
	}
}
