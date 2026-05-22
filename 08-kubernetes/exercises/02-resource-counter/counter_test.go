//go:build exercise

package counter

import (
	"context"
	"errors"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fakeCluster returns canned per-namespace data.
type fakeCluster struct {
	namespaces []string

	podsByNS    map[string]int
	depsByNS    map[string]int
	svcsByNS    map[string]int

	errPodsByNS map[string]error
}

func (f *fakeCluster) ListNamespaces(_ context.Context, _ metav1.ListOptions) (*corev1.NamespaceList, error) {
	out := &corev1.NamespaceList{}
	for _, n := range f.namespaces {
		out.Items = append(out.Items, corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: n}})
	}
	return out, nil
}

func (f *fakeCluster) ListPods(_ context.Context, ns string, _ metav1.ListOptions) (*corev1.PodList, error) {
	if err, ok := f.errPodsByNS[ns]; ok {
		return nil, err
	}
	out := &corev1.PodList{}
	for i := 0; i < f.podsByNS[ns]; i++ {
		out.Items = append(out.Items, corev1.Pod{})
	}
	return out, nil
}

func (f *fakeCluster) ListDeployments(_ context.Context, ns string, _ metav1.ListOptions) (*appsv1.DeploymentList, error) {
	out := &appsv1.DeploymentList{}
	for i := 0; i < f.depsByNS[ns]; i++ {
		out.Items = append(out.Items, appsv1.Deployment{})
	}
	return out, nil
}

func (f *fakeCluster) ListServices(_ context.Context, ns string, _ metav1.ListOptions) (*corev1.ServiceList, error) {
	out := &corev1.ServiceList{}
	for i := 0; i < f.svcsByNS[ns]; i++ {
		out.Items = append(out.Items, corev1.Service{})
	}
	return out, nil
}

func TestCount_TallyPerNamespace(t *testing.T) {
	f := &fakeCluster{
		namespaces: []string{"default", "kube-system"},
		podsByNS:   map[string]int{"default": 3, "kube-system": 7},
		depsByNS:   map[string]int{"default": 1, "kube-system": 4},
		svcsByNS:   map[string]int{"default": 2, "kube-system": 5},
	}

	rows, err := Count(context.Background(), f)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(rows))
	}
	if rows[0].Namespace != "default" || rows[0].Pods != 3 || rows[0].Deployments != 1 || rows[0].Services != 2 {
		t.Errorf("row[0] = %+v, want {default 3 1 2}", rows[0])
	}
	if rows[1].Namespace != "kube-system" || rows[1].Pods != 7 || rows[1].Deployments != 4 || rows[1].Services != 5 {
		t.Errorf("row[1] = %+v, want {kube-system 7 4 5}", rows[1])
	}
}

func TestCount_EmptyCluster(t *testing.T) {
	f := &fakeCluster{}

	rows, err := Count(context.Background(), f)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("rows = %+v, want empty", rows)
	}
}

func TestCount_NamespaceWithNoResources(t *testing.T) {
	f := &fakeCluster{
		namespaces: []string{"empty"},
		// no entries in any of the byNS maps — defaults to 0
	}

	rows, err := Count(context.Background(), f)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(rows))
	}
	if rows[0].Namespace != "empty" || rows[0].Pods != 0 || rows[0].Deployments != 0 || rows[0].Services != 0 {
		t.Errorf("row[0] = %+v, want {empty 0 0 0}", rows[0])
	}
}

func TestCount_PreservesNamespaceOrder(t *testing.T) {
	f := &fakeCluster{namespaces: []string{"zulu", "alpha", "mike"}}

	rows, err := Count(context.Background(), f)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	got := []string{rows[0].Namespace, rows[1].Namespace, rows[2].Namespace}
	if got[0] != "zulu" || got[1] != "alpha" || got[2] != "mike" {
		t.Errorf("namespace order = %v, want [zulu alpha mike] (no sort)", got)
	}
}

func TestCount_ListErrorPropagates(t *testing.T) {
	stub := errors.New("forbidden")
	f := &fakeCluster{
		namespaces:  []string{"good", "bad"},
		podsByNS:    map[string]int{"good": 1},
		errPodsByNS: map[string]error{"bad": stub},
	}

	_, err := Count(context.Background(), f)
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}
