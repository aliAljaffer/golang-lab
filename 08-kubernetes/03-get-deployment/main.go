// 03-get-deployment — get a specific Deployment and print its replica count.
//
// What this example proves:
//   - `clientset.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})`
//     is the "fetch one resource by name" shape. Symmetric to List but takes
//     a name and returns a single object (not a List).
//   - Deployments live in `apps/v1` — that's the API group. Pods live in
//     `core/v1` (`CoreV1()` is just the unprefixed core group).
//   - `dep.Status.Replicas` vs `dep.Spec.Replicas` is the canonical
//     desired-vs-actual split. *Spec* is what you asked for; *Status* is
//     what the controller has reconciled.
//   - `errors.IsNotFound(err)` is how you distinguish "deployment doesn't
//     exist" from "API server is broken."
//
// Run:
//
//	go run . --namespace kube-system --name coredns
//	go run . --namespace default --name does-not-exist  # IsNotFound path
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	ns := flag.String("namespace", "default", "namespace")
	name := flag.String("name", "", "deployment name (required)")
	flag.Parse()

	if *name == "" {
		fmt.Fprintln(os.Stderr, "--name is required")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "load config:", err)
		os.Exit(1)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "build clientset:", err)
		os.Exit(1)
	}

	// TODO: dep, err := clientset.AppsV1().Deployments(*ns).Get(ctx, *name, metav1.GetOptions{})
	// TODO: if apierrors.IsNotFound(err) { fmt.Fprintln(os.Stderr, "not found"); os.Exit(3) }
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "get:", err); os.Exit(1) }

	// TODO: desired := int32(0)
	// TODO: if dep.Spec.Replicas != nil { desired = *dep.Spec.Replicas }
	// TODO: fmt.Printf("%s/%s — desired=%d ready=%d available=%d\n",
	// TODO:     dep.Namespace, dep.Name, desired, dep.Status.ReadyReplicas, dep.Status.AvailableReplicas)

	_ = clientset
	_ = ctx
	_ = ns
	_ = metav1.GetOptions{}
	_ = apierrors.IsNotFound
}

func loadConfig() (*rest.Config, error) {
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}
	path := os.Getenv("KUBECONFIG")
	if path == "" {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, ".kube", "config")
	}
	return clientcmd.BuildConfigFromFlags("", path)
}
