// 07-webhook-receiver — verify HMAC-SHA256 signatures on incoming webhooks.
//
// Webhooks from GitHub, Slack, Stripe all sign their payloads with HMAC. The
// receiver re-computes the HMAC using a shared secret and compares it to the
// signature header in **constant time** (hmac.Equal — never bytes.Equal).
//
// Reference (GitHub format):
//   X-Hub-Signature-256: sha256=<hex>
//   computed over the raw request body using the webhook secret.
//
// Run:
//   WEBHOOK_SECRET=swordfish go run .
//   ./send-signed.sh   # see README for a sample curl that signs locally
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// verifySignature returns true iff `header` (e.g. "sha256=<hex>") is a valid
// HMAC-SHA256 of `body` keyed by `secret`. Constant-time comparison via hmac.Equal.
func verifySignature(secret []byte, body []byte, header string) bool {
	// TODO: header should start with "sha256=" — strip that prefix.
	// TODO: decode the hex tail into want []byte.
	// TODO: mac := hmac.New(sha256.New, secret); mac.Write(body); got := mac.Sum(nil)
	// TODO: return hmac.Equal(got, want)
	_ = strings.TrimPrefix
	_ = hex.DecodeString
	_ = hmac.New
	_ = sha256.New
	return false
}

func handleWebhook(secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		// TODO: sig := r.Header.Get("X-Hub-Signature-256")
		// TODO: if !verifySignature(secret, body, sig) { http.Error(w, "bad signature", 401); return }
		// TODO: fmt.Fprintln(w, "ok")

		_ = io.ReadAll
		http.Error(w, "TODO", http.StatusNotImplemented)
	}
}

func main() {
	secret := []byte(os.Getenv("WEBHOOK_SECRET"))
	if len(secret) == 0 {
		fmt.Fprintln(os.Stderr, "WEBHOOK_SECRET is required")
		os.Exit(2)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhook", handleWebhook(secret))

	log.Fatal(http.ListenAndServe(":8080", mux))
}
