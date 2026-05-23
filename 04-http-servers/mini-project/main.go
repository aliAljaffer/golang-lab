// webhook-runner — a tiny CI runner.
//
// Spec (from ../PLAN.md):
//   - POST /webhook receives a signed payload: { "job": "<name>" }
//   - Header X-Hub-Signature-256: sha256=<hex>  (HMAC-SHA256 of body, GitHub-style)
//   - YAML config maps job names to a shell command line
//   - On valid signature + known job, runs the command and replies
//     { "job": "...", "exit_code": N, "output": "..." } with output truncated.
//   - On bad signature -> 401. On unknown job -> 404.
//   - SIGINT/SIGTERM triggers graceful shutdown: drain in-flight runs, then exit.
//
// Testable surface (top of the file). cobra wiring is at the bottom.
//
// Run:
//   WEBHOOK_SECRET=swordfish go run . -c webhooks.yaml
package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	defaultMaxBody       = 1 << 20 // 1 MiB
	defaultMaxOutput     = 4 << 10 // 4 KiB of captured stdout+stderr
	defaultShutdownGrace = 10 * time.Second
)

// Config maps job names to commands.
type Config struct {
	Jobs map[string]Job `yaml:"jobs"`
}

// Job is a single command to run, as an argv slice (no shell interpolation).
type Job struct {
	Command []string `yaml:"command"`
}

// WebhookRequest is the JSON body of POST /webhook.
type WebhookRequest struct {
	Job string `json:"job"`
}

// WebhookResponse is the success reply.
type WebhookResponse struct {
	Job      string `json:"job"`
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output"`
}

// LoadConfig reads a YAML file at `path` and parses it into a Config.
// A nil or empty `Jobs` map is treated as an error — pointless to run without jobs.
func LoadConfig(path string) (Config, error) {
	// TODO: read the YAML and reject obviously-useless configs (no jobs at
	//   all, or a job declared with an empty Command). The test pins the
	//   exact error path for each — read it before you write.

	_ = yaml.Unmarshal
	return Config{}, fmt.Errorf("LoadConfig: not implemented")
}

// VerifyHMAC validates a GitHub-style "sha256=<hex>" signature header.
// Compares in constant time via hmac.Equal.
func VerifyHMAC(secret, body []byte, header string) bool {
	// TODO: parse the "sha256=<hex>" header and compare against the
	//   HMAC-SHA256 of body under secret. CRITICAL: compare with hmac.Equal,
	//   not bytes.Equal or "==". Constant-time comparison is the entire point
	//   of an HMAC — short-circuiting on a length/byte mismatch leaks the
	//   prefix of the right signature to a network observer.

	_ = strings.HasPrefix
	_ = hex.DecodeString
	_ = hmac.New
	_ = sha256.New
	return false
}

// runJob runs `j.Command` with the given context. Returns the process exit code
// and combined stdout+stderr output, truncated to maxOutput bytes.
//
// Exit codes:
//   - 0   on success
//   - >0  on process failure (from *exec.ExitError)
//   - The returned err is non-nil only for failures *outside* the process's
//     control (couldn't spawn, ctx cancelled before start, etc.).
func runJob(ctx context.Context, j Job, maxOutput int) (exitCode int, output string, err error) {
	// TODO: spawn the command (exec.CommandContext so ctx-cancel kills it),
	//   capture stdout+stderr into one buffer, and truncate to maxOutput.
	//   The interesting bit is the error split documented above: an
	//   *exec.ExitError (the program ran and exited non-zero) is NOT a Go
	//   error from runJob's perspective — caller wants the exit code, not
	//   an HTTP 500. Only failures-to-spawn / ctx-cancels return a non-nil
	//   err from runJob.

	_ = bytes.NewBuffer
	_ = exec.CommandContext
	return 0, "", fmt.Errorf("runJob: not implemented")
}

// newHandler returns the /webhook handler. Exported indirectly via newServer so
// the cobra wiring stays slim and tests can drive the handler with httptest.
func newHandler(cfg Config, secret []byte, maxOutput int) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhook", func(w http.ResponseWriter, r *http.Request) {
		// TODO: drive the request through: cap body size, verify HMAC, parse,
		//   dispatch to runJob, return JSON. The status-code mapping is
		//   pinned by the tests:
		//     - bad/missing signature -> 401
		//     - malformed JSON / oversized body -> 400
		//     - unknown job -> 404
		//     - runJob err (couldn't spawn) -> 500
		//     - success (incl. job that exited non-zero) -> 200 with exit_code
		//   The "exited non-zero is still 200" line trips people up — the
		//   runner did its job; the command failed.

		_ = io.ReadAll
		_ = json.Unmarshal
		http.Error(w, "newHandler: not implemented", http.StatusNotImplemented)
	})
	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func newRootCmd() *cobra.Command {
	var (
		configPath string
		addr       string
	)
	cmd := &cobra.Command{
		Use:   "webhook-runner",
		Short: "HMAC-verified webhook receiver that runs configured commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig(configPath)
			if err != nil {
				return err
			}
			secret := []byte(os.Getenv("WEBHOOK_SECRET"))
			if len(secret) == 0 {
				return errors.New("WEBHOOK_SECRET env var is required")
			}

			srv := &http.Server{
				Addr:              addr,
				Handler:           newHandler(cfg, secret, defaultMaxOutput),
				ReadHeaderTimeout: 5 * time.Second,
			}

			// TODO: graceful shutdown — run ListenAndServe on a goroutine,
			//   wait on a signal channel, then Shutdown with the grace
			//   deadline. http.ErrServerClosed coming out of ListenAndServe
			//   is the *expected* path after Shutdown, not an error to
			//   propagate.

			_ = signal.Notify
			_ = syscall.SIGTERM
			_ = srv
			return errors.New("webhook-runner: not implemented")
		},
	}
	cmd.Flags().StringVarP(&configPath, "config", "c", "webhooks.yaml", "path to YAML job config")
	cmd.Flags().StringVar(&addr, "addr", ":8080", "listen address")
	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
