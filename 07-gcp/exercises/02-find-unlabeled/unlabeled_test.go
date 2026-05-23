//go:build exercise

package unlabeled

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// fakeCompute returns a canned flat instance list.
type fakeCompute struct {
	instances []InstanceSummary
	err       error
}

func (f *fakeCompute) AggregatedListInstances(_ context.Context, _ string) ([]InstanceSummary, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.instances, nil
}

func inst(name, zone string, labels map[string]string) InstanceSummary {
	return InstanceSummary{Name: name, Zone: zone, Labels: labels}
}

func TestFindUnlabeled_FlagsMissingKey(t *testing.T) {
	f := &fakeCompute{instances: []InstanceSummary{
		inst("vm-good", "zones/us-central1-a", map[string]string{"env": "prod"}),
		inst("vm-bad", "zones/us-central1-a", nil), // no labels at all
	}}

	got, err := FindUnlabeled(context.Background(), f, "p", "env")
	if err != nil {
		t.Fatalf("FindUnlabeled: %v", err)
	}
	if len(got) != 1 || got[0] != "vm-bad" {
		t.Errorf("got %v, want [vm-bad]", got)
	}
}

func TestFindUnlabeled_EmptyValueCountsAsPresent(t *testing.T) {
	// The KEY existing is what counts. An empty value (env="") is a valid
	// label — kubectl, gcloud, and most production checks treat it that way.
	f := &fakeCompute{instances: []InstanceSummary{
		inst("vm-empty", "zones/us-central1-a", map[string]string{"env": ""}),
	}}

	got, err := FindUnlabeled(context.Background(), f, "p", "env")
	if err != nil {
		t.Fatalf("FindUnlabeled: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %v, want [] (key exists, value empty but valid)", got)
	}
}

func TestFindUnlabeled_PreservesAPIOrder(t *testing.T) {
	f := &fakeCompute{instances: []InstanceSummary{
		inst("vm-a", "zones/us-central1-a", nil),
		inst("vm-b", "zones/us-central1-a", map[string]string{"env": "prod"}),
		inst("vm-c", "zones/europe-west1-b", nil),
	}}

	got, err := FindUnlabeled(context.Background(), f, "p", "env")
	if err != nil {
		t.Fatalf("FindUnlabeled: %v", err)
	}
	if strings.Join(got, ",") != "vm-a,vm-c" {
		t.Errorf("got %v, want [vm-a vm-c] (preserve API order across zones)", got)
	}
}

func TestFindUnlabeled_AllInstancesLabeled(t *testing.T) {
	f := &fakeCompute{instances: []InstanceSummary{
		inst("vm-1", "zones/us-central1-a", map[string]string{"env": "prod", "owner": "ali"}),
		inst("vm-2", "zones/us-central1-a", map[string]string{"env": "staging"}),
	}}

	got, err := FindUnlabeled(context.Background(), f, "p", "env")
	if err != nil {
		t.Fatalf("FindUnlabeled: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %v, want [] (every instance has the env label)", got)
	}
}

func TestFindUnlabeled_RejectsEmptyKey(t *testing.T) {
	f := &fakeCompute{}

	_, err := FindUnlabeled(context.Background(), f, "p", "")
	if err == nil {
		t.Error("FindUnlabeled(\"\"): err = nil, want validation error")
	}
}

func TestFindUnlabeled_PropagatesError(t *testing.T) {
	stub := errors.New("permission denied")
	f := &fakeCompute{err: stub}

	_, err := FindUnlabeled(context.Background(), f, "p", "env")
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}
