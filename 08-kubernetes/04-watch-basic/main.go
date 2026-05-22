// 04-watch-basic — raw Watch() over pod events.
//
// What this example proves:
//   - `Pods(ns).Watch(ctx, opts)` returns a `watch.Interface`. Its `ResultChan()`
//     emits `watch.Event{Type, Object}` until the watch closes or context cancels.
//   - Event types: `watch.Added`, `watch.Modified`, `watch.Deleted`, `watch.Error`.
//   - Raw watches have a fatal flaw — they CAN and WILL die. The API server
//     rotates connections; etcd compaction can invalidate your resource version.
//     When that happens you get `watch.Error` (often `StatusReasonGone`) and
//     you have to restart the watch yourself.
//   - This is why informers exist (see example 05). Use raw Watch only for
//     short-lived "until first event then quit" patterns, or when you genuinely
//     want every event and are willing to re-subscribe on errors.
//
// Run:
//
//	go run . --namespace default
//	# in another terminal:
//	kubectl run pinger --image=busybox --restart=Never -- sh -c 'sleep 3'
//	kubectl delete pod pinger
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	ns := flag.String("namespace", "default", "namespace to watch (empty for all)")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
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

	// TODO: w, err := clientset.CoreV1().Pods(*ns).Watch(ctx, metav1.ListOptions{})
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "watch:", err); os.Exit(1) }
	// TODO: defer w.Stop()

	// TODO: for event := range w.ResultChan() {
	// TODO:     pod, ok := event.Object.(*corev1.Pod)
	// TODO:     if !ok {
	// TODO:         fmt.Printf("non-pod event: %T\n", event.Object)
	// TODO:         continue
	// TODO:     }
	// TODO:     fmt.Printf("%-8s %s/%s phase=%s\n", event.Type, pod.Namespace, pod.Name, pod.Status.Phase)
	// TODO: }
	// TODO: fmt.Println("watch channel closed")

	_ = clientset
	_ = ctx
	_ = ns
	_ = metav1.ListOptions{}
	_ = corev1.Pod{}
	_ = watch.Added
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
