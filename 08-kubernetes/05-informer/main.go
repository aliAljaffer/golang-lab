// 05-informer — a SharedInformer with Add/Update/Delete handlers.
//
// What this example proves:
//   - `informers.NewSharedInformerFactory(clientset, resync)` builds a factory
//     that hands out one informer per resource type.
//   - The factory's underlying machinery: ListAndWatch with auto-reconnect,
//     local cache (a thread-safe store keyed by namespace/name), and event
//     handlers that fire AFTER the cache is updated.
//   - `factory.Start(stopCh)` kicks off goroutines; `factory.WaitForCacheSync`
//     blocks until each registered informer has its initial List complete.
//     Don't read from the cache before sync — it'll be empty and you'll do
//     surprising things.
//   - The "shared" in SharedInformer means multiple handlers share ONE watch
//     connection. Build five things that want to know about pods? Still one
//     watch.
//
// Run:
//
//	go run . --namespace default
//	# in another terminal:
//	kubectl run pinger --image=busybox --restart=Never -- sh -c 'sleep 60'
//	kubectl label pod pinger ohai=world  # triggers Update
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
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	ns := flag.String("namespace", "", "namespace to watch (empty == all namespaces)")
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

	// TODO: opt := informers.WithNamespace(*ns)
	// TODO: factory := informers.NewSharedInformerFactoryWithOptions(clientset, 30*time.Second, opt)
	// TODO: podInformer := factory.Core().V1().Pods()

	// TODO: _, err = podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	// TODO:     AddFunc: func(obj interface{}) {
	// TODO:         pod := obj.(*corev1.Pod)
	// TODO:         fmt.Printf("ADD     %s/%s phase=%s\n", pod.Namespace, pod.Name, pod.Status.Phase)
	// TODO:     },
	// TODO:     UpdateFunc: func(oldObj, newObj interface{}) {
	// TODO:         oldPod, newPod := oldObj.(*corev1.Pod), newObj.(*corev1.Pod)
	// TODO:         if oldPod.Status.Phase == newPod.Status.Phase {
	// TODO:             return  // periodic resync — don't spam on no-op events
	// TODO:         }
	// TODO:         fmt.Printf("UPDATE  %s/%s %s -> %s\n", newPod.Namespace, newPod.Name, oldPod.Status.Phase, newPod.Status.Phase)
	// TODO:     },
	// TODO:     DeleteFunc: func(obj interface{}) {
	// TODO:         pod, ok := obj.(*corev1.Pod)
	// TODO:         if !ok { /* tombstone case; see README */ return }
	// TODO:         fmt.Printf("DELETE  %s/%s\n", pod.Namespace, pod.Name)
	// TODO:     },
	// TODO: })
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "register handler:", err); os.Exit(1) }

	// TODO: factory.Start(ctx.Done())
	// TODO: if !cache.WaitForCacheSync(ctx.Done(), podInformer.Informer().HasSynced) {
	// TODO:     fmt.Fprintln(os.Stderr, "cache sync timed out")
	// TODO:     os.Exit(1)
	// TODO: }
	// TODO: fmt.Println("informer running; ctrl-c to exit")
	// TODO: <-ctx.Done()

	_ = clientset
	_ = ctx
	_ = ns
	_ = time.Second
	_ = informers.NewSharedInformerFactory
	_ = cache.ResourceEventHandlerFuncs{}
	_ = corev1.Pod{}
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
