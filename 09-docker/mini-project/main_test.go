//go:build exercise

package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
)

// ---- isUntagged ------------------------------------------------------------

func TestIsUntagged_NilTagsIsUntagged(t *testing.T) {
	if !isUntagged(image.Summary{ID: "sha256:a", RepoTags: nil}) {
		t.Error("isUntagged(nil RepoTags) = false, want true")
	}
}

func TestIsUntagged_DanglingPlaceholderIsUntagged(t *testing.T) {
	if !isUntagged(image.Summary{ID: "sha256:a", RepoTags: []string{"<none>:<none>"}}) {
		t.Error(`isUntagged(RepoTags=["<none>:<none>"]) = false, want true`)
	}
}

func TestIsUntagged_RealTagIsTagged(t *testing.T) {
	if isUntagged(image.Summary{ID: "sha256:a", RepoTags: []string{"alpine:3"}}) {
		t.Error(`isUntagged(["alpine:3"]) = true, want false`)
	}
}

// ---- Plan ------------------------------------------------------------------

func img(id string, tags []string, created time.Time) image.Summary {
	return image.Summary{ID: id, RepoTags: tags, Created: created.Unix()}
}

func ctr(imageID string) container.Summary {
	return container.Summary{ImageID: imageID}
}

var fixedNow = time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC)

func TestPlan_RemovesUntagged(t *testing.T) {
	imgs := []image.Summary{
		img("sha256:tagged", []string{"alpine:3"}, fixedNow),
		img("sha256:dangling", []string{"<none>:<none>"}, fixedNow),
	}
	got := Plan(imgs, nil, Policy{RemoveUntagged: true}, fixedNow)

	want := []string{"sha256:dangling"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Plan = %v, want %v", got, want)
	}
}

func TestPlan_RemovesOldImages(t *testing.T) {
	imgs := []image.Summary{
		img("sha256:fresh", []string{"a:1"}, fixedNow.Add(-1*time.Hour)),
		img("sha256:stale", []string{"b:1"}, fixedNow.Add(-30*24*time.Hour)),
	}
	got := Plan(imgs, nil, Policy{RemoveOlderThan: 7 * 24 * time.Hour}, fixedNow)

	want := []string{"sha256:stale"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Plan = %v, want %v", got, want)
	}
}

func TestPlan_KeepsImagesWithContainerReferences(t *testing.T) {
	imgs := []image.Summary{
		img("sha256:used", nil, fixedNow),
		img("sha256:unused", nil, fixedNow),
	}
	containers := []container.Summary{ctr("sha256:used")}

	got := Plan(imgs, containers, Policy{RemoveUnused: true}, fixedNow)

	want := []string{"sha256:unused"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Plan = %v, want %v (referenced images must be kept)", got, want)
	}
}

func TestPlan_PoliciesAreOred(t *testing.T) {
	// Untagged OR older-than-7d. Stale-tagged hits old, dangling hits untagged.
	imgs := []image.Summary{
		img("sha256:keep", []string{"alpine:3"}, fixedNow),
		img("sha256:dangling", nil, fixedNow),
		img("sha256:stale", []string{"old:1"}, fixedNow.Add(-30*24*time.Hour)),
	}
	got := Plan(imgs, nil, Policy{RemoveUntagged: true, RemoveOlderThan: 7 * 24 * time.Hour}, fixedNow)

	want := []string{"sha256:dangling", "sha256:stale"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Plan = %v, want %v", got, want)
	}
}

func TestPlan_EmptyPolicyEmptyResult(t *testing.T) {
	imgs := []image.Summary{img("sha256:a", nil, fixedNow)}
	got := Plan(imgs, nil, Policy{}, fixedNow)
	if len(got) != 0 {
		t.Errorf("Plan with no policy = %v, want empty", got)
	}
}

func TestPlan_OutputIsSorted(t *testing.T) {
	imgs := []image.Summary{
		img("sha256:zulu", nil, fixedNow),
		img("sha256:alpha", nil, fixedNow),
		img("sha256:mike", nil, fixedNow),
	}
	got := Plan(imgs, nil, Policy{RemoveUntagged: true}, fixedNow)

	want := []string{"sha256:alpha", "sha256:mike", "sha256:zulu"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Plan = %v, want %v (lexicographic order)", got, want)
	}
}

// ---- Sync ------------------------------------------------------------------

// fakeDocker is a hand-rolled stand-in for the DockerAPI interface. Counts
// ImageRemove calls so the dry-run test can prove zero mutations.
type fakeDocker struct {
	images       []image.Summary
	containers   []container.Summary
	listErr      error
	removeErr    error
	removeCalls  atomic.Int32
	removedSeen  []string
}

func (f *fakeDocker) ImageList(_ context.Context, _ image.ListOptions) ([]image.Summary, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.images, nil
}

func (f *fakeDocker) ContainerList(_ context.Context, _ container.ListOptions) ([]container.Summary, error) {
	return f.containers, nil
}

func (f *fakeDocker) ImageRemove(_ context.Context, id string, _ image.RemoveOptions) ([]image.DeleteResponse, error) {
	f.removeCalls.Add(1)
	f.removedSeen = append(f.removedSeen, id)
	if f.removeErr != nil {
		return nil, f.removeErr
	}
	return []image.DeleteResponse{{Deleted: id}}, nil
}

func TestSync_HappyPathRemovesImages(t *testing.T) {
	f := &fakeDocker{
		images: []image.Summary{
			img("sha256:dangling", nil, fixedNow),
			img("sha256:tagged", []string{"alpine:3"}, fixedNow),
		},
	}
	got, err := Sync(context.Background(), f, Policy{RemoveUntagged: true}, fixedNow, io.Discard)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if strings.Join(got, ",") != "sha256:dangling" {
		t.Errorf("returned ids = %v, want [sha256:dangling]", got)
	}
	if f.removeCalls.Load() != 1 {
		t.Errorf("ImageRemove called %d times, want 1", f.removeCalls.Load())
	}
}

func TestSync_DryRunMakesNoRemoveCalls(t *testing.T) {
	f := &fakeDocker{
		images: []image.Summary{
			img("sha256:a", nil, fixedNow),
			img("sha256:b", nil, fixedNow),
		},
	}
	var buf bytes.Buffer
	got, err := Sync(context.Background(), f, Policy{RemoveUntagged: true, DryRun: true}, fixedNow, &buf)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("returned ids = %v, want 2 entries", got)
	}
	if f.removeCalls.Load() != 0 {
		t.Fatalf("ImageRemove called %d times in dry-run, want 0", f.removeCalls.Load())
	}
	// Dry-run output should mention "would remove" (not "removed").
	if !strings.Contains(buf.String(), "would remove") {
		t.Errorf("dry-run output = %q, want to contain 'would remove'", buf.String())
	}
	if strings.Contains(buf.String(), "removed sha256:") {
		t.Errorf("dry-run output = %q, must not contain 'removed <id>'", buf.String())
	}
}

func TestSync_ListErrorPropagates(t *testing.T) {
	stub := errors.New("daemon down")
	f := &fakeDocker{listErr: stub}

	_, err := Sync(context.Background(), f, Policy{RemoveUntagged: true}, fixedNow, io.Discard)
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}

func TestSync_RemoveErrorPropagates(t *testing.T) {
	stub := errors.New("image in use")
	f := &fakeDocker{
		images:    []image.Summary{img("sha256:a", nil, fixedNow)},
		removeErr: stub,
	}
	_, err := Sync(context.Background(), f, Policy{RemoveUntagged: true}, fixedNow, io.Discard)
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v wrapped", err, stub)
	}
}

func TestSync_ForceFlagThreadsThrough(t *testing.T) {
	// This test catches the "did you remember to pass p.Force into RemoveOptions?"
	// regression. We wrap ImageRemove to capture the opts.
	var seen image.RemoveOptions
	f := &captureRemoveFake{
		base: &fakeDocker{images: []image.Summary{img("sha256:a", nil, fixedNow)}},
		seen: &seen,
	}
	_, err := Sync(context.Background(), f, Policy{RemoveUntagged: true, Force: true}, fixedNow, io.Discard)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if !seen.Force {
		t.Error("ImageRemove was called with Force=false, want true (flag must thread through)")
	}
}

// captureRemoveFake wraps fakeDocker and stashes the RemoveOptions of the last
// ImageRemove call.
type captureRemoveFake struct {
	base *fakeDocker
	seen *image.RemoveOptions
}

func (c *captureRemoveFake) ImageList(ctx context.Context, opts image.ListOptions) ([]image.Summary, error) {
	return c.base.ImageList(ctx, opts)
}

func (c *captureRemoveFake) ContainerList(ctx context.Context, opts container.ListOptions) ([]container.Summary, error) {
	return c.base.ContainerList(ctx, opts)
}

func (c *captureRemoveFake) ImageRemove(ctx context.Context, id string, opts image.RemoveOptions) ([]image.DeleteResponse, error) {
	*c.seen = opts
	return c.base.ImageRemove(ctx, id, opts)
}
