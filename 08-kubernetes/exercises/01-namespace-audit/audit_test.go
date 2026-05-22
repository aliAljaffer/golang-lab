//go:build exercise

package audit

import (
	"context"
	"errors"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fakeNS returns a canned list. errOnList primes a List failure.
type fakeNS struct {
	items     []corev1.Namespace
	errOnList error
}

func (f *fakeNS) List(_ context.Context, _ metav1.ListOptions) (*corev1.NamespaceList, error) {
	if f.errOnList != nil {
		return nil, f.errOnList
	}
	return &corev1.NamespaceList{Items: f.items}, nil
}

func nsObj(name string, labels map[string]string) corev1.Namespace {
	return corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels},
	}
}

func TestAudit_FlagsNamespacesMissingLabel(t *testing.T) {
	f := &fakeNS{items: []corev1.Namespace{
		nsObj("default", map[string]string{"owner": "platform"}),
		nsObj("rogue", nil),
		nsObj("rogue2", map[string]string{"other": "thing"}),
	}}

	got, err := Audit(context.Background(), f, "owner")
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	want := []string{"rogue", "rogue2"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Audit = %v, want %v", got, want)
	}
}

func TestAudit_AllNamespacesHaveLabel(t *testing.T) {
	f := &fakeNS{items: []corev1.Namespace{
		nsObj("a", map[string]string{"owner": "x"}),
		nsObj("b", map[string]string{"owner": "y"}),
	}}

	got, err := Audit(context.Background(), f, "owner")
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Audit = %v, want empty", got)
	}
}

func TestAudit_EmptyValueCountsAsPresent(t *testing.T) {
	// Same semantics as kubectl: `kubectl label ns/x owner=` sets owner="".
	// The key is present, so the namespace is NOT flagged.
	f := &fakeNS{items: []corev1.Namespace{
		nsObj("blank-owner", map[string]string{"owner": ""}),
	}}

	got, err := Audit(context.Background(), f, "owner")
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Audit = %v, want empty (empty value still counts as present)", got)
	}
}

func TestAudit_PreservesAPIOrder(t *testing.T) {
	f := &fakeNS{items: []corev1.Namespace{
		nsObj("zulu", nil),
		nsObj("alpha", nil),
		nsObj("mike", nil),
	}}

	got, err := Audit(context.Background(), f, "owner")
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	want := []string{"zulu", "alpha", "mike"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Audit = %v, want %v (order from API, no sort)", got, want)
	}
}

func TestAudit_ListErrorPropagates(t *testing.T) {
	stub := errors.New("forbidden")
	f := &fakeNS{errOnList: stub}

	_, err := Audit(context.Background(), f, "owner")
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}
