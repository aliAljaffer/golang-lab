//go:build exercise

package rollrestart

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// fakeDep records every call.
type fakeDep struct {
	calls []patchCall
	err   error
}

type patchCall struct {
	name string
	pt   types.PatchType
	data []byte
}

func (f *fakeDep) Patch(_ context.Context, name string, pt types.PatchType, data []byte, _ metav1.PatchOptions, _ ...string) (*appsv1.Deployment, error) {
	f.calls = append(f.calls, patchCall{name: name, pt: pt, data: append([]byte(nil), data...)})
	if f.err != nil {
		return nil, f.err
	}
	return &appsv1.Deployment{}, nil
}

func TestRollingRestart_StrategicMergePatch(t *testing.T) {
	f := &fakeDep{}
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)

	if err := RollingRestart(context.Background(), f, "default", "web", now); err != nil {
		t.Fatalf("RollingRestart: %v", err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("Patch called %d times, want 1", len(f.calls))
	}
	if f.calls[0].pt != types.StrategicMergePatchType {
		t.Errorf("patch type = %v, want StrategicMergePatchType", f.calls[0].pt)
	}
}

func TestRollingRestart_PatchTargetsCorrectName(t *testing.T) {
	f := &fakeDep{}
	if err := RollingRestart(context.Background(), f, "default", "web", time.Now()); err != nil {
		t.Fatalf("RollingRestart: %v", err)
	}
	if f.calls[0].name != "web" {
		t.Errorf("Patch(name) = %q, want \"web\"", f.calls[0].name)
	}
}

func TestRollingRestart_BodyContainsRFC3339Timestamp(t *testing.T) {
	f := &fakeDep{}
	now := time.Date(2026, 5, 22, 12, 30, 45, 0, time.UTC)
	wantStamp := now.Format(time.RFC3339) // "2026-05-22T12:30:45Z"

	if err := RollingRestart(context.Background(), f, "default", "web", now); err != nil {
		t.Fatalf("RollingRestart: %v", err)
	}

	// Parse the JSON patch body and check the annotation is set with the right value.
	var doc map[string]any
	if err := json.Unmarshal(f.calls[0].data, &doc); err != nil {
		t.Fatalf("patch body is not valid JSON: %v\nbody=%s", err, f.calls[0].data)
	}
	spec, _ := doc["spec"].(map[string]any)
	tmpl, _ := spec["template"].(map[string]any)
	meta, _ := tmpl["metadata"].(map[string]any)
	annots, _ := meta["annotations"].(map[string]any)
	gotStamp, _ := annots[RestartedAtKey].(string)
	if gotStamp != wantStamp {
		t.Errorf("annotation %s = %q, want %q\nfull body=%s", RestartedAtKey, gotStamp, wantStamp, f.calls[0].data)
	}
}

func TestRollingRestart_PropagatesError(t *testing.T) {
	stub := errors.New("forbidden")
	f := &fakeDep{err: stub}

	err := RollingRestart(context.Background(), f, "default", "web", time.Now())
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}

func TestRollingRestart_TwoCallsProduceDistinctTimestamps(t *testing.T) {
	f := &fakeDep{}
	t1 := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	t2 := t1.Add(time.Minute)

	if err := RollingRestart(context.Background(), f, "default", "web", t1); err != nil {
		t.Fatalf("first: %v", err)
	}
	if err := RollingRestart(context.Background(), f, "default", "web", t2); err != nil {
		t.Fatalf("second: %v", err)
	}
	if string(f.calls[0].data) == string(f.calls[1].data) {
		t.Errorf("two patches produced identical bodies — timestamp didn't refresh:\n  %s\n  %s", f.calls[0].data, f.calls[1].data)
	}
}
