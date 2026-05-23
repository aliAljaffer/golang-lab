// kube-events-to-slack — watches every Event via an informer, filters by
// severity + namespace + age, dedupes per <namespace>/<kind>/<name>:<reason>,
// and posts the result to a Slack incoming-webhook URL (or stdout in
// --dry-run mode).
//
// The testable surface lives in filter.go / dedup.go / sink.go / format.go /
// run.go. This file is just the cobra wiring + kubeconfig loader and is not
// covered by the exercise tests.
//
// Run:
//
//	go run . --dry-run --severities Warning
//	go run . --webhook-url https://hooks.slack.com/services/... --namespace kube-system
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type runOpts struct {
	Namespaces []string
	WebhookURL string
	Severities []string
	Cooldown   time.Duration
	MaxAge     time.Duration
	DryRun     bool
	Kubeconfig string
}

func newRootCmd() *cobra.Command {
	var opts runOpts
	cmd := &cobra.Command{
		Use:   "kube-events-to-slack",
		Short: "Watch k8s Events, filter+dedup, post to a Slack webhook",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadConfig(opts.Kubeconfig)
			if err != nil {
				return fmt.Errorf("load kubeconfig: %w", err)
			}
			clientset, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				return fmt.Errorf("build clientset: %w", err)
			}

			filter := Filter{
				Severities: stringSet(opts.Severities),
				Namespaces: stringSet(opts.Namespaces),
				MaxAge:     opts.MaxAge,
				Now:        time.Now,
			}

			var sink Sink
			if opts.DryRun || opts.WebhookURL == "" {
				sink = &StdoutSink{Out: cmd.OutOrStdout()}
			} else {
				sink = &WebhookSink{URL: opts.WebhookURL, Client: http.DefaultClient, MaxRetries: 3}
			}
			deduper := NewDeduper(opts.Cooldown)

			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()
			return Run(ctx, clientset, filter, deduper, sink, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringSliceVar(&opts.Namespaces, "namespace", nil, "namespace allow-list (repeatable; empty == all)")
	cmd.Flags().StringVar(&opts.WebhookURL, "webhook-url", "", "Slack incoming-webhook URL (empty or --dry-run == stdout)")
	cmd.Flags().StringSliceVar(&opts.Severities, "severities", []string{"Warning"}, "event severities to alert on (Normal,Warning)")
	cmd.Flags().DurationVar(&opts.Cooldown, "cooldown", 5*time.Minute, "dedup cooldown per key")
	cmd.Flags().DurationVar(&opts.MaxAge, "max-age", time.Hour, "skip events older than this")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print alerts as JSON lines instead of POSTing")
	cmd.Flags().StringVar(&opts.Kubeconfig, "kubeconfig", "", "path to kubeconfig (default: in-cluster, then $KUBECONFIG, then ~/.kube/config)")
	return cmd
}

func stringSet(items []string) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	out := make(map[string]bool, len(items))
	for _, s := range items {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		out[s] = true
	}
	return out
}

func loadConfig(explicit string) (*rest.Config, error) {
	if explicit != "" {
		return clientcmd.BuildConfigFromFlags("", explicit)
	}
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
