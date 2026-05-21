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
	// TODO: data, err := os.ReadFile(path)
	// TODO: var c Config; yaml.Unmarshal(data, &c)
	// TODO: if len(c.Jobs) == 0 { return Config{}, fmt.Errorf("no jobs in %s", path) }
	// TODO: for name, j := range c.Jobs { if len(j.Command) == 0 { return Config{}, fmt.Errorf("job %q has empty command", name) } }

	_ = yaml.Unmarshal
	return Config{}, fmt.Errorf("LoadConfig: not implemented")
}

// VerifyHMAC validates a GitHub-style "sha256=<hex>" signature header.
// Compares in constant time via hmac.Equal.
func VerifyHMAC(secret, body []byte, header string) bool {
	// TODO: const prefix = "sha256="
	// TODO: if !strings.HasPrefix(header, prefix) { return false }
	// TODO: want, err := hex.DecodeString(header[len(prefix):]); if err != nil { return false }
	// TODO: mac := hmac.New(sha256.New, secret); mac.Write(body); got := mac.Sum(nil)
	// TODO: return hmac.Equal(got, want)

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
	// TODO: if len(j.Command) == 0 { return 0, "", errors.New("empty command") }
	// TODO: cmd := exec.CommandContext(ctx, j.Command[0], j.Command[1:]...)
	// TODO: var buf bytes.Buffer; cmd.Stdout = &buf; cmd.Stderr = &buf
	// TODO: runErr := cmd.Run()
	// TODO: out := buf.String(); if len(out) > maxOutput { out = out[:maxOutput] }
	// TODO:
	//   switch e := runErr.(type) {
	//   case nil:                return 0, out, nil
	//   case *exec.ExitError:    return e.ExitCode(), out, nil
	//   default:                 return 0, out, runErr
	//   }

	_ = bytes.NewBuffer
	_ = exec.CommandContext
	return 0, "", fmt.Errorf("runJob: not implemented")
}

// newHandler returns the /webhook handler. Exported indirectly via newServer so
// the cobra wiring stays slim and tests can drive the handler with httptest.
func newHandler(cfg Config, secret []byte, maxOutput int) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhook", func(w http.ResponseWriter, r *http.Request) {
		// TODO: r.Body = http.MaxBytesReader(w, r.Body, defaultMaxBody)
		// TODO: body, err := io.ReadAll(r.Body); if err != nil { http.Error(w, err.Error(), 400); return }
		// TODO: if !VerifyHMAC(secret, body, r.Header.Get("X-Hub-Signature-256")) {
		//          http.Error(w, "bad signature", http.StatusUnauthorized); return
		//       }
		// TODO: var req WebhookRequest
		//       if err := json.Unmarshal(body, &req); err != nil { http.Error(w, err.Error(), 400); return }
		// TODO: job, ok := cfg.Jobs[req.Job]
		//       if !ok { http.Error(w, "unknown job", http.StatusNotFound); return }
		// TODO: exit, out, err := runJob(r.Context(), job, maxOutput)
		//       if err != nil { http.Error(w, err.Error(), 500); return }
		// TODO: writeJSON(w, 200, WebhookResponse{Job: req.Job, ExitCode: exit, Output: out})

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

			// TODO: run srv.ListenAndServe() in a goroutine; forward non-ErrServerClosed errors to errCh.
			// TODO: signal.Notify on SIGINT + SIGTERM; on signal, srv.Shutdown(ctx with defaultShutdownGrace).
			// TODO: return whichever channel fires first.

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
