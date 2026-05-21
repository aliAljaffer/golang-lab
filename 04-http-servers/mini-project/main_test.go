//go:build exercise && !windows

package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testSecret = "swordfish"

// sign builds a GitHub-style "sha256=<hex>" header for body, keyed by secret.
func sign(secret, body []byte) string {
	m := hmac.New(sha256.New, secret)
	m.Write(body)
	return "sha256=" + hex.EncodeToString(m.Sum(nil))
}

// writeYAML drops `body` into a temp file and returns its path.
func writeYAML(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "webhooks.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	return path
}

// doSigned posts body to /webhook on baseURL with a valid HMAC.
func doSigned(t *testing.T, baseURL string, body []byte) *http.Response {
	t.Helper()
	req, err := http.NewRequest("POST", baseURL+"/webhook", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("X-Hub-Signature-256", sign([]byte(testSecret), body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	return resp
}

func TestLoadConfig_ParsesYAML(t *testing.T) {
	path := writeYAML(t, `
jobs:
  echo:
    command: ["echo", "hi"]
  failing:
    command: ["sh", "-c", "exit 7"]
`)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Jobs) != 2 {
		t.Fatalf("got %d jobs, want 2: %+v", len(cfg.Jobs), cfg.Jobs)
	}
	if got := cfg.Jobs["echo"].Command; len(got) != 2 || got[0] != "echo" || got[1] != "hi" {
		t.Errorf("echo command = %v, want [echo hi]", got)
	}
}

func TestLoadConfig_RejectsEmpty(t *testing.T) {
	path := writeYAML(t, "jobs: {}\n")
	if _, err := LoadConfig(path); err == nil {
		t.Fatal("LoadConfig: want error on empty jobs map, got nil")
	}
}

func TestVerifyHMAC(t *testing.T) {
	secret := []byte(testSecret)
	body := []byte(`{"job":"echo"}`)
	good := sign(secret, body)

	cases := []struct {
		name   string
		header string
		body   []byte
		want   bool
	}{
		{"good", good, body, true},
		{"missing prefix", strings.TrimPrefix(good, "sha256="), body, false},
		{"tampered body", good, []byte(`{"job":"other"}`), false},
		{"empty header", "", body, false},
		{"bad hex", "sha256=not-hex-zz", body, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := VerifyHMAC(secret, c.body, c.header); got != c.want {
				t.Errorf("VerifyHMAC(%q) = %v, want %v", c.name, got, c.want)
			}
		})
	}
}

// startTestServer spins up an http.Server bound to a random port with the given handler.
// It returns the base URL and a teardown that calls srv.Close().
func startTestServer(t *testing.T, h http.Handler) (string, *http.Server) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	srv := &http.Server{Handler: h, ReadHeaderTimeout: 2 * time.Second}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })
	return "http://" + ln.Addr().String(), srv
}

func TestWebhook_RejectsBadSignature(t *testing.T) {
	cfg := Config{Jobs: map[string]Job{
		"echo": {Command: []string{"echo", "hi"}},
	}}
	h := newHandler(cfg, []byte(testSecret), defaultMaxOutput)
	base, _ := startTestServer(t, h)

	body := []byte(`{"job":"echo"}`)

	req, _ := http.NewRequest("POST", base+"/webhook", bytes.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", "sha256=deadbeef")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}

	req2, _ := http.NewRequest("POST", base+"/webhook", bytes.NewReader(body))
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusUnauthorized {
		t.Errorf("unsigned: status = %d, want 401", resp2.StatusCode)
	}
}

func TestWebhook_UnknownJob(t *testing.T) {
	cfg := Config{Jobs: map[string]Job{
		"echo": {Command: []string{"echo", "hi"}},
	}}
	h := newHandler(cfg, []byte(testSecret), defaultMaxOutput)
	base, _ := startTestServer(t, h)

	resp := doSigned(t, base, []byte(`{"job":"nope"}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestWebhook_RunsJob_CapturesExitCode(t *testing.T) {
	cfg := Config{Jobs: map[string]Job{
		"ok":   {Command: []string{"sh", "-c", "echo ran && exit 0"}},
		"fail": {Command: []string{"sh", "-c", "echo nope >&2 && exit 7"}},
	}}
	h := newHandler(cfg, []byte(testSecret), defaultMaxOutput)
	base, _ := startTestServer(t, h)

	t.Run("success", func(t *testing.T) {
		resp := doSigned(t, base, []byte(`{"job":"ok"}`))
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		var got WebhookResponse
		if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.ExitCode != 0 {
			t.Errorf("ExitCode = %d, want 0", got.ExitCode)
		}
		if !strings.Contains(got.Output, "ran") {
			t.Errorf("Output = %q, want it to contain %q", got.Output, "ran")
		}
	})

	t.Run("failure", func(t *testing.T) {
		resp := doSigned(t, base, []byte(`{"job":"fail"}`))
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("status = %d, want 200 (the *job* failed, not the request)", resp.StatusCode)
		}
		var got WebhookResponse
		if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.ExitCode != 7 {
			t.Errorf("ExitCode = %d, want 7", got.ExitCode)
		}
		if !strings.Contains(got.Output, "nope") {
			t.Errorf("Output = %q, want it to contain stderr 'nope'", got.Output)
		}
	})
}

func TestWebhook_TruncatesOutput(t *testing.T) {
	// Generate ~2KiB of output; cap at 256 bytes.
	const maxOut = 256
	cfg := Config{Jobs: map[string]Job{
		"big": {Command: []string{"sh", "-c", `for i in $(seq 1 200); do printf '0123456789'; done`}},
	}}
	h := newHandler(cfg, []byte(testSecret), maxOut)
	base, _ := startTestServer(t, h)

	resp := doSigned(t, base, []byte(`{"job":"big"}`))
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var got WebhookResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode: %v (body=%s)", err, body)
	}
	if len(got.Output) != maxOut {
		t.Errorf("len(Output) = %d, want %d (truncation)", len(got.Output), maxOut)
	}
}

func TestServer_GracefulShutdownDrainsInFlight(t *testing.T) {
	// Job that takes ~300ms; we'll start a request, then call Shutdown, and
	// expect the request to complete with 200 before Shutdown returns.
	cfg := Config{Jobs: map[string]Job{
		"slow": {Command: []string{"sh", "-c", "sleep 0.3 && echo done"}},
	}}
	h := newHandler(cfg, []byte(testSecret), defaultMaxOutput)
	base, srv := startTestServer(t, h)

	body := []byte(`{"job":"slow"}`)
	respCh := make(chan *http.Response, 1)
	errCh := make(chan error, 1)
	go func() {
		req, _ := http.NewRequest("POST", base+"/webhook", bytes.NewReader(body))
		req.Header.Set("X-Hub-Signature-256", sign([]byte(testSecret), body))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			errCh <- err
			return
		}
		respCh <- resp
	}()

	// Give the request a moment to reach the handler.
	time.Sleep(80 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown returned error: %v (should drain cleanly)", err)
	}

	select {
	case resp := <-respCh:
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Errorf("in-flight response status = %d, want 200", resp.StatusCode)
		}
	case err := <-errCh:
		t.Fatalf("in-flight request errored — shutdown did not drain: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("in-flight request did not complete after Shutdown returned")
	}
}

// keep imports honest even when individual tests are commented out
var _ = fmt.Sprintf
