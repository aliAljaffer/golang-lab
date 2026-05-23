// crashloop-alert — watch every pod via an informer; alert when one enters
// CrashLoopBackOff. Dedup by <namespace>/<name> for a configurable cooldown.
//
// Testable surface (top of file). The cobra/informer wiring is at the bottom.
//
// Run:
//
//	go run . --cooldown 30s
//	go run . --webhook https://hooks.example/alerts --cooldown 5m
//	go run . --namespace kube-system
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/spf13/cobra"
)

// Alert is what one detection emits. Marshalled to JSON for the webhook sink.
type Alert struct {
	Namespace string    `json:"namespace"`
	Name      string    `json:"name"`
	Container string    `json:"container"`
	Reason    string    `json:"reason"`
	At        time.Time `json:"at"`
}

// Sink is anything that can deliver an Alert. Stdout, HTTP, log aggregator —
// implementations are below. The handler closes over the configured Sink.
type Sink interface {
	Send(ctx context.Context, alert Alert) error
}

// StdoutSink writes one JSON-encoded line per alert to Out.
type StdoutSink struct {
	Out io.Writer
}

// Send writes a JSON line. Concurrency-safe if Out is.
func (s *StdoutSink) Send(_ context.Context, alert Alert) error {
	// TODO: emit one JSON line per call. ctx is ignored intentionally —
	//   writes to s.Out shouldn't be cancellable.
	return errors.New("StdoutSink.Send not implemented")
}

// WebhookSink POSTs JSON-encoded alerts to URL. Client is the http.Client to
// use; nil means http.DefaultClient.
type WebhookSink struct {
	URL    string
	Client *http.Client
}

// Send posts a single alert. Non-2xx is an error. Network errors propagate.
func (s *WebhookSink) Send(ctx context.Context, alert Alert) error {
	// TODO: POST the JSON-marshalled alert to s.URL via ctx-aware request.
	//   Don't forget Content-Type. Non-2xx is an error containing the status
	//   code so callers can tell timeouts apart from a misconfigured webhook.
	return errors.New("WebhookSink.Send not implemented")
}

// IsCrashLooping returns true if any container in pod is in Waiting with
// reason "CrashLoopBackOff". Returns false for nil pods, finished pods, etc.
//
// The reason string is the public contract — kubelet sets it whenever the
// CrashLoop back-off timer kicks in. It does NOT appear on Pending pods
// that haven't started yet — those are "ContainerCreating" or "PodInitializing".
func IsCrashLooping(pod *corev1.Pod) bool {
	// TODO: scan the pod's container statuses for a Waiting state with
	//   Reason == "CrashLoopBackOff". Don't skip InitContainerStatuses —
	//   init containers crashloop too, and missing them means quiet failures
	//   on pods that never make it past init.
	_ = corev1.PodStatus{}
	return false
}

// CrashLoopingContainer returns the first crashlooping container name, or "".
// Useful for the Alert's Container field.
func CrashLoopingContainer(pod *corev1.Pod) string {
	// TODO: same scan as IsCrashLooping, but return the container name on
	//   first hit. Empty string for "none" — used directly in the Alert.
	return ""
}

// Deduper rate-limits alerts per key. Use NewDeduper to build one.
type Deduper struct {
	mu       sync.Mutex
	cooldown time.Duration
	lastSent map[string]time.Time
	Now      func() time.Time // injected for tests; defaults to time.Now
}

// NewDeduper returns a Deduper with the given cooldown. Now defaults to time.Now.
func NewDeduper(cooldown time.Duration) *Deduper {
	return &Deduper{
		cooldown: cooldown,
		lastSent: map[string]time.Time{},
		Now:      time.Now,
	}
}

// ShouldAlert returns true if no alert for `key` was sent within the cooldown.
// Side effect: records the alert time on a true return so the NEXT call within
// the cooldown returns false.
func (d *Deduper) ShouldAlert(key string) bool {
	// TODO: read+update lastSent under d.mu — informers fire from several
	//   goroutines. On a true return you also have to RECORD now, otherwise
	//   the cooldown does nothing. d.Now is the clock seam — tests use it.
	return false
}

// newPodHandler returns a cache.ResourceEventHandler that runs detect+dedup+send.
// Send errors are logged to errOut but do not stop processing.
func newPodHandler(d *Deduper, sink Sink, errOut io.Writer) cache.ResourceEventHandler {
	check := func(obj interface{}) {
		// TODO: detect -> dedup -> send. The informer hands you interface{},
		//   so type-assert to *corev1.Pod first (return on unexpected types).
		//   Sink errors go to errOut; don't return them, or the informer will
		//   drop the item.
	}
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    check,
		UpdateFunc: func(_, obj interface{}) { check(obj) },
	}
}

// Run starts the informer factory and blocks until ctx is cancelled. ns="" watches
// every namespace; otherwise restricts to ns. Returns nil on clean shutdown.
func Run(ctx context.Context, clientset kubernetes.Interface, ns string, d *Deduper, sink Sink, errOut io.Writer) error {
	// TODO: build a SharedInformerFactory (use the WithNamespace option only
	//   when ns is non-empty), attach newPodHandler to the Pods informer,
	//   Start the factory, and block on ctx.Done(). WaitForCacheSync is the
	//   load-bearing call — skipping it makes every existing crashlooping
	//   pod fire as a "new" alert on every restart.
	_ = informers.NewSharedInformerFactory
	_ = cache.WaitForCacheSync
	return errors.New("Run not implemented")
}

// helper: silence unused-import lints during scaffolding.
var (
	_ = bytes.NewReader
	_ = json.Marshal
)

// ---- cobra wiring (not unit-tested) ----------------------------------------

type runOpts struct {
	Namespace string
	Cooldown  time.Duration
	Webhook   string
}

func newRootCmd() *cobra.Command {
	var opts runOpts
	cmd := &cobra.Command{
		Use:   "crashloop-alert",
		Short: "Watch pods cluster-wide, alert on CrashLoopBackOff with dedup",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			clientset, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				return fmt.Errorf("build clientset: %w", err)
			}

			var sink Sink
			if opts.Webhook == "" {
				sink = &StdoutSink{Out: cmd.OutOrStdout()}
			} else {
				sink = &WebhookSink{URL: opts.Webhook, Client: http.DefaultClient}
			}
			d := NewDeduper(opts.Cooldown)

			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()
			return Run(ctx, clientset, opts.Namespace, d, sink, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&opts.Namespace, "namespace", "", "namespace to watch (empty == all)")
	cmd.Flags().DurationVar(&opts.Cooldown, "cooldown", 5*time.Minute, "dedup cooldown per pod")
	cmd.Flags().StringVar(&opts.Webhook, "webhook", "", "POST alerts here as JSON (empty == stdout)")
	return cmd
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

func main() {
	if err := newRootCmd().ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
