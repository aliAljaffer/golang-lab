// Package audit finds namespaces missing a required label.
//
// Exercise surface:
//
//	type NamespaceAPI interface { List }
//	func Audit(ctx, api NamespaceAPI, requiredLabel string) ([]string, error)
//
// Tests pass a fake NamespaceAPI. There is no `main.go` — wire your own CLI if you want.
package audit

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceAPI is the slice of kubernetes.Interface this package needs. A real
// `*kubernetes.Clientset` satisfies it via `.CoreV1().Namespaces()`.
//
// (Defining a one-method interface here is the production pattern: callers
// pass `clientset.CoreV1().Namespaces()` directly, and tests pass a fake.)
type NamespaceAPI interface {
	List(ctx context.Context, opts metav1.ListOptions) (*corev1.NamespaceList, error)
}

// Audit lists every namespace and returns the names of those whose labels
// do NOT include requiredLabel as a key. The label value is irrelevant —
// only the key's presence is checked.
//
// Order: preserve the order the API returned (don't sort).
// Empty value counts as "label is present" (kubectl-equivalent contract).
//
// Hints:
//   - api.List(ctx, metav1.ListOptions{}) returns *corev1.NamespaceList with .Items
//   - Each item has .ObjectMeta.Labels (a map[string]string, can be nil)
//   - Check `_, ok := ns.Labels[requiredLabel]; !ok` — that's the "missing key" path
func Audit(ctx context.Context, api NamespaceAPI, requiredLabel string) ([]string, error) {
	// TODO: implement.
	_ = corev1.NamespaceList{}
	return nil, errors.New("Audit not implemented")
}
