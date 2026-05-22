// 02-list-pods — list pods in a namespace and filter by label.
//
// What this example proves:
//   - `clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{...})`
//     is the standard "list resources" shape.
//   - `metav1.ListOptions{LabelSelector: "app=foo"}` filters server-side — the
//     API server applies the selector, not the client. Saves bandwidth.
//   - The returned `*corev1.PodList` has `.Items []corev1.Pod`. Pod status
//     lives in `.Status.Phase` ("Pending", "Running", "Succeeded", "Failed").
//
// Run:
//
//	go run . --namespace default
//	go run . --namespace kube-system --selector k8s-app=kube-dns
//	go run . --namespace ""           # empty == all namespaces (v1.PodsGetter convention)
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	ns := flag.String("namespace", "default", "namespace to list pods in (\"\" for all namespaces)")
	sel := flag.String("selector", "", "label selector, e.g. app=nginx")
	flag.Parse()

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

	// TODO: pods, err := clientset.CoreV1().Pods(*ns).List(ctx, metav1.ListOptions{LabelSelector: *sel})
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "list pods:", err); os.Exit(1) }

	// TODO: for _, p := range pods.Items {
	// TODO:     fmt.Printf("%-30s %-10s %s\n", p.Name, p.Status.Phase, p.Spec.NodeName)
	// TODO: }

	_ = clientset
	_ = ctx
	_ = ns
	_ = sel
	_ = metav1.ListOptions{}
	_ = corev1.PodList{}
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
