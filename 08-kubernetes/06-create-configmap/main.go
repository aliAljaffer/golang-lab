// 06-create-configmap — create a ConfigMap programmatically.
//
// What this example proves:
//   - `Create(ctx, obj, metav1.CreateOptions{})` is the universal write shape.
//     Symmetric to `Get` and `List`. Symmetric across resource types.
//   - You build the resource as a regular Go struct — `&corev1.ConfigMap{...}`.
//     Fields with embedded types (`TypeMeta`, `ObjectMeta`) live at the top
//     level due to embedding; you set `Name`, `Namespace`, `Labels` directly.
//   - The API server fills in `ResourceVersion`, `UID`, `CreationTimestamp`,
//     so don't bother setting them.
//   - `apierrors.IsAlreadyExists(err)` is how you handle the conflict from
//     a second call. The idiomatic "upsert" is Get → if NotFound Create,
//     else Update. (Or use Apply server-side — outside this example's scope.)
//
// Run:
//
//	go run . --namespace default --name demo-config
//	kubectl get configmap demo-config -o yaml
//	kubectl delete configmap demo-config
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	ns := flag.String("namespace", "default", "namespace")
	name := flag.String("name", "demo-config", "configmap name")
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

	// TODO: cm := &corev1.ConfigMap{
	// TODO:     ObjectMeta: metav1.ObjectMeta{
	// TODO:         Name:      *name,
	// TODO:         Namespace: *ns,
	// TODO:         Labels:    map[string]string{"managed-by": "08-create-configmap"},
	// TODO:     },
	// TODO:     Data: map[string]string{
	// TODO:         "greeting": "hello",
	// TODO:         "created":  time.Now().Format(time.RFC3339),
	// TODO:     },
	// TODO: }

	// TODO: out, err := clientset.CoreV1().ConfigMaps(*ns).Create(ctx, cm, metav1.CreateOptions{})
	// TODO: if apierrors.IsAlreadyExists(err) {
	// TODO:     fmt.Fprintln(os.Stderr, "already exists — delete or change --name")
	// TODO:     os.Exit(3)
	// TODO: }
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "create:", err); os.Exit(1) }
	// TODO: fmt.Printf("created %s/%s (uid=%s)\n", out.Namespace, out.Name, out.UID)

	_ = clientset
	_ = ctx
	_ = ns
	_ = name
	_ = corev1.ConfigMap{}
	_ = metav1.CreateOptions{}
	_ = apierrors.IsAlreadyExists
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
