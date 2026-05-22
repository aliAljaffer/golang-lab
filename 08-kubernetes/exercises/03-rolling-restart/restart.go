// Package rollrestart triggers a rolling restart of a Deployment by patching
// the pod template's annotations with a fresh timestamp.
//
// Why this works: `kubectl rollout restart` is implemented exactly this way.
// Touching the pod template's PodSpec (even an annotation under it) bumps the
// Deployment's generation, which triggers the controller to roll out new pods.
// We're patching `spec.template.metadata.annotations.kubectl.kubernetes.io/restartedAt`
// — same key kubectl uses, so this plays nicely with `kubectl rollout status`.
//
// Exercise surface:
//
//	type DeploymentAPI interface { Patch }
//	func RollingRestart(ctx, api DeploymentAPI, ns, name string, now time.Time) error
package rollrestart

import (
	"context"
	"errors"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// DeploymentAPI is the slice of clientset.AppsV1().Deployments(ns) this package needs.
//
// The real clientset's method signature for Patch is the same — that's why
// the interface is small. Tests pass a fake that records every call.
type DeploymentAPI interface {
	Patch(
		ctx context.Context,
		name string,
		pt types.PatchType,
		data []byte,
		opts metav1.PatchOptions,
		subresources ...string,
	) (*appsv1.Deployment, error)
}

// RestartedAtKey is the annotation key kubectl uses for `rollout restart`.
const RestartedAtKey = "kubectl.kubernetes.io/restartedAt"

// RollingRestart patches the Deployment ns/name to set
// `spec.template.metadata.annotations[RestartedAtKey] = now.Format(time.RFC3339)`,
// which causes the Deployment controller to roll its pods.
//
// Use a strategic-merge patch (types.StrategicMergePatchType). The patch body
// is a tiny JSON document of the shape:
//
//	{ "spec": { "template": { "metadata": { "annotations": { "kubectl.kubernetes.io/restartedAt": "2026-05-22T00:00:00Z" } } } } }
//
// Hints:
//   - Build the body with `fmt.Sprintf` or `json.Marshal` of a nested map[string]any.
//   - api.Patch(ctx, name, types.StrategicMergePatchType, body, metav1.PatchOptions{})
//   - Now is injected so tests can pin a known timestamp; default to time.Now in callers.
func RollingRestart(ctx context.Context, api DeploymentAPI, ns, name string, now time.Time) error {
	// TODO: implement.
	_ = types.StrategicMergePatchType
	_ = RestartedAtKey
	return errors.New("RollingRestart not implemented")
}
