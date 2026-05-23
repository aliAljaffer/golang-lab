// deploy-bot — fetch a GitHub release artifact, build a Docker image from
// it, run the container, and verify it answers /healthz before reporting.
//
// The testable surface lives in github.go / download.go / build.go /
// runctr.go / health.go / run.go. This file is just the cobra wiring +
// SDK construction and is not covered by the exercise tests.
//
// Run:
//
//	go run . deploy alialjaffer/example v1.2.3 --gh-token "$GH_TOKEN"
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

type rootOpts struct {
	GHToken         string
	HostPort        int
	ContainerPort   int
	HealthPath      string
	HealthTimeout   time.Duration
	HealthInterval  time.Duration
	Env             []string
	KeepContainer   bool
	DryRun          bool
}

func newRootCmd() *cobra.Command {
	var opts rootOpts
	cmd := &cobra.Command{
		Use:   "deploy-bot",
		Short: "Fetch a GitHub release, build it, run it, probe /healthz",
	}
	deployCmd := &cobra.Command{
		Use:   "deploy <owner/repo> <tag>",
		Short: "Deploy a tagged release of a GitHub repo as a local container",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			owner, repo, ok := splitOwnerRepo(args[0])
			if !ok {
				return fmt.Errorf("first arg %q must be in owner/repo form", args[0])
			}
			tag := args[1]

			httpClient := &http.Client{Timeout: 30 * time.Second}
			fetcher := &GHReleaseFetcher{
				HTTPClient: httpClient,
				Token:      opts.GHToken,
				BaseURL:    "https://api.github.com",
			}
			downloader := &HTTPDownloader{
				Client: httpClient,
				Token:  opts.GHToken,
			}

			dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				return fmt.Errorf("build docker client: %w", err)
			}
			defer dockerCli.Close()

			builder := &DockerBuilder{Inner: &dockerBuildAdapter{Client: dockerCli}}
			runner := &DockerRunner{Inner: &dockerRunAdapter{Client: dockerCli}}
			health := &HTTPHealthChecker{
				Client:   httpClient,
				Interval: opts.HealthInterval,
				Sleep:    ctxSleep,
			}

			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			runCtx, runCancel := context.WithTimeout(ctx, opts.HealthTimeout)
			defer runCancel()

			report, err := Run(runCtx, fetcher, downloader, builder, runner, health, Opts{
				Owner:         owner,
				Repo:          repo,
				Tag:           tag,
				HealthPath:    opts.HealthPath,
				HostPort:      opts.HostPort,
				ContainerPort: opts.ContainerPort,
				Env:           opts.Env,
				KeepContainer: opts.KeepContainer,
				DryRun:        opts.DryRun,
			})
			fmt.Fprintf(cmd.OutOrStdout(), "report: %+v\n", report)
			return err
		},
	}
	deployCmd.Flags().StringVar(&opts.GHToken, "gh-token", os.Getenv("GH_TOKEN"), "GitHub API token (defaults to $GH_TOKEN)")
	deployCmd.Flags().IntVar(&opts.HostPort, "host-port", 8080, "host port to publish")
	deployCmd.Flags().IntVar(&opts.ContainerPort, "container-port", 8080, "container port to expose")
	deployCmd.Flags().StringVar(&opts.HealthPath, "health-path", "/healthz", "path probed on http://localhost:<host-port>")
	deployCmd.Flags().DurationVar(&opts.HealthTimeout, "health-timeout", 30*time.Second, "total time the pipeline has to go green")
	deployCmd.Flags().DurationVar(&opts.HealthInterval, "health-interval", time.Second, "delay between health probes")
	deployCmd.Flags().StringSliceVar(&opts.Env, "env", nil, "KEY=VALUE env to set on the container (repeatable)")
	deployCmd.Flags().BoolVar(&opts.KeepContainer, "keep-container", false, "do NOT remove the container on health failure")
	deployCmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print the pipeline plan but do not call Docker")

	cmd.AddCommand(deployCmd)
	return cmd
}

// splitOwnerRepo parses "owner/repo" into its two halves. Returns ok=false
// if the input is not exactly two non-empty slash-separated segments.
func splitOwnerRepo(s string) (owner, repo string, ok bool) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// ctxSleep is a ctx-cancellable sleep used by HTTPHealthChecker between
// probes. Lives in main.go because health.go injects it; tests substitute
// their own (e.g. fastSleep that ignores the duration).
func ctxSleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func main() {
	if err := newRootCmd().ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
