// Package counter tallies pods, deployments, and services per namespace.
//
// Exercise surface:
//
//	type ClusterAPI interface { ListNamespaces; ListPods; ListDeployments; ListServices }
//	type Row struct { Namespace string; Pods, Deployments, Services int }
//	func Count(ctx, api ClusterAPI) ([]Row, error)
//
// Tests pass a fake. There is no `main.go` — wire your own CLI if you want.
package counter

import (
	"context"
	"errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterAPI is the slice of kubernetes.Interface this package needs.
//
// Defining four methods on one interface (rather than four one-method
// interfaces) is a judgment call. Pick whichever you prefer; both are
// idiomatic. The PLAN-recommended shape is four methods, one interface.
type ClusterAPI interface {
	ListNamespaces(ctx context.Context, opts metav1.ListOptions) (*corev1.NamespaceList, error)
	ListPods(ctx context.Context, ns string, opts metav1.ListOptions) (*corev1.PodList, error)
	ListDeployments(ctx context.Context, ns string, opts metav1.ListOptions) (*appsv1.DeploymentList, error)
	ListServices(ctx context.Context, ns string, opts metav1.ListOptions) (*corev1.ServiceList, error)
}

// Row is one namespace's tally.
type Row struct {
	Namespace   string
	Pods        int
	Deployments int
	Services    int
}

// Count returns one Row per namespace, in the order the API returned the namespaces.
// On error from any list call, returns the partial rows accumulated so far
// PLUS the error (caller can choose to display partial results or not).
//
// Hints:
//   - ListNamespaces gives you the .Items to iterate.
//   - For each namespace, call ListPods/Deployments/Services with ns=name.
//   - len(list.Items) is the tally.
func Count(ctx context.Context, api ClusterAPI) ([]Row, error) {
	// TODO: implement.
	_ = corev1.NamespaceList{}
	_ = appsv1.DeploymentList{}
	return nil, errors.New("Count not implemented")
}
