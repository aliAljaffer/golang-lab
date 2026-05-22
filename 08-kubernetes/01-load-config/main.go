// 01-load-config — load a kubeconfig out-of-cluster and print the server version.
//
// What this example proves:
//   - `clientcmd.BuildConfigFromFlags("", kubeconfigPath)` reads a kubeconfig
//     file and returns a *rest.Config.
//   - `kubernetes.NewForConfig(cfg)` builds a typed clientset that exposes one
//     interface per API group (`CoreV1()`, `AppsV1()`, etc.).
//   - The Discovery client hits `/version` — the cheapest "is this cluster
//     reachable?" call you can make. Use it as your smoke test.
//
// In-cluster vs out-of-cluster:
//   - Inside a pod: use `rest.InClusterConfig()` (reads the ServiceAccount
//     token + CA from the pod's filesystem).
//   - Outside (your laptop): load `~/.kube/config` via clientcmd.
//   - The pattern below tries in-cluster first, falls back to kubeconfig —
//     this is the canonical "works both places" boilerplate.
//
// Run:
//
//	go run .
//	KUBECONFIG=$HOME/.kube/config go run .
//
// Requires a reachable cluster (minikube/kind/colima) for the version call.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// TODO: cfg, err := loadConfig()
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "load config:", err); os.Exit(1) }

	// TODO: clientset, err := kubernetes.NewForConfig(cfg)
	// TODO: if err != nil { ... os.Exit(1) }

	// TODO: ver, err := clientset.Discovery().ServerVersion()
	// TODO: if err != nil { ... os.Exit(1) }
	// TODO: fmt.Printf("server: %s.%s (%s)\n", ver.Major, ver.Minor, ver.GitVersion)

	_ = ctx
	_ = kubernetes.NewForConfig
	_ = os.Exit
	_ = fmt.Println
}

// loadConfig tries in-cluster first, then falls back to a kubeconfig file
// at $KUBECONFIG or ~/.kube/config. This is the canonical "works in both
// places" pattern — write it once, copy it for every k8s tool you build.
func loadConfig() (*rest.Config, error) {
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}
	path := os.Getenv("KUBECONFIG")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("locate home dir: %w", err)
		}
		path = filepath.Join(home, ".kube", "config")
	}
	return clientcmd.BuildConfigFromFlags("", path)
}
