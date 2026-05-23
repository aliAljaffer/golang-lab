// s3-log-shipper — tail local logs, batch+gzip them, upload to S3.
//
// The testable surface lives in tail.go / batch.go / upload.go / run.go.
// This file is just the cobra wiring + S3 client construction and is not
// covered by the exercise tests.
//
// Run:
//
//	go run . --path /var/log/myapp.log --bucket my-archive --key-prefix logs/myapp
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

type runOpts struct {
	Paths         []string
	Bucket        string
	KeyPrefix     string
	Region        string
	MaxBatchBytes int
	MaxBatchAge   time.Duration
	OffsetDir     string
	MaxRetries    int
}

func newRootCmd() *cobra.Command {
	var opts runOpts
	cmd := &cobra.Command{
		Use:   "s3-log-shipper",
		Short: "Tail local logs, batch+gzip, upload to S3",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.Bucket == "" {
				return errors.New("--bucket is required")
			}
			if len(opts.Paths) == 0 {
				return errors.New("at least one --path is required")
			}

			loaders := []func(*config.LoadOptions) error{}
			if opts.Region != "" {
				loaders = append(loaders, config.WithRegion(opts.Region))
			}
			cfg, err := config.LoadDefaultConfig(cmd.Context(), loaders...)
			if err != nil {
				return fmt.Errorf("load aws config: %w", err)
			}
			client := s3.NewFromConfig(cfg)

			hostname, _ := os.Hostname()
			store := &FileOffsetStore{Dir: opts.OffsetDir}

			tailers := make([]*Tailer, 0, len(opts.Paths))
			for _, p := range opts.Paths {
				tailers = append(tailers, &Tailer{
					Path:         p,
					Store:        store,
					PollInterval: 200 * time.Millisecond,
				})
			}

			batcher := &Batcher{
				MaxBytes: opts.MaxBatchBytes,
				MaxAge:   opts.MaxBatchAge,
				Hostname: hostname,
				Now:      time.Now,
			}

			uploader := &S3Uploader{
				Inner:       &s3ClientAdapter{Client: client},
				Bucket:      opts.Bucket,
				MaxRetries:  opts.MaxRetries,
				BaseBackoff: 500 * time.Millisecond,
				Now:         time.Now,
				Sleep:       ctxSleep,
			}

			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()
			return Run(ctx, tailers, batcher, uploader, opts.KeyPrefix, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringSliceVar(&opts.Paths, "path", nil, "log file to tail (repeatable)")
	cmd.Flags().StringVar(&opts.Bucket, "bucket", "", "destination S3 bucket")
	cmd.Flags().StringVar(&opts.KeyPrefix, "key-prefix", "", "prefix prepended to every uploaded object key")
	cmd.Flags().StringVar(&opts.Region, "region", "", "AWS region (empty == use SDK default chain)")
	cmd.Flags().IntVar(&opts.MaxBatchBytes, "max-batch-bytes", 1<<20, "flush a batch when raw line bytes meet this threshold")
	cmd.Flags().DurationVar(&opts.MaxBatchAge, "max-batch-age", 30*time.Second, "flush a batch this long after its first line")
	cmd.Flags().StringVar(&opts.OffsetDir, "offset-dir", "", "directory for offset sidecar files (empty == sibling to each --path)")
	cmd.Flags().IntVar(&opts.MaxRetries, "max-retries", 3, "retry count for transient S3 errors")
	return cmd
}

// ctxSleep is a ctx-cancellable sleep used by S3Uploader for backoff. Lives
// in main.go because Run / the uploader inject it; tests substitute their own.
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
