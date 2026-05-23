// 01-adc-and-project — load Application Default Credentials, inspect which
// source fired, and resolve the active project ID.
//
// What this example proves:
//   - `google.FindDefaultCredentials(ctx, scope)` walks the ADC chain:
//       GOOGLE_APPLICATION_CREDENTIALS env var → `gcloud auth
//       application-default login` user creds → GCE/Cloud Run metadata server
//   - Project ID is a separate dimension. ADC creds know which token to mint,
//     not which project to act against. It comes from (in order):
//       GOOGLE_CLOUD_PROJECT env var → the active gcloud config → the JSON key
//       file's `project_id` field (when present) → metadata server (when on GCP).
//   - Credentials are resolved lazily — finding them does NOT mint a token.
//     The actual auth call happens when you make a service call.
//
// Run:
//
//	go run .
//	GOOGLE_APPLICATION_CREDENTIALS=/path/to/sa.json go run .
//	GOOGLE_CLOUD_PROJECT=my-project go run .
//
// Requires either a logged-in gcloud or a service-account JSON.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2/google"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// PART 1 — find the credentials. cloud-platform is the broadest scope; any
	// service-specific token request the SDK makes later is a subset of this.
	// TODO: creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "find default creds:", err); os.Exit(1) }

	// PART 2 — show the project ID baked into the creds (when a JSON key was
	// used) and the env override.
	// TODO: fmt.Println("creds.ProjectID:", creds.ProjectID)
	// TODO: fmt.Println("GOOGLE_CLOUD_PROJECT:", os.Getenv("GOOGLE_CLOUD_PROJECT"))

	// PART 3 — force a token mint so you can see auth actually fire. (In a
	// real client, the SDK calls this for you on the first request.)
	// TODO: tok, err := creds.TokenSource.Token()
	// TODO: if err != nil { fmt.Fprintln(os.Stderr, "token:", err); os.Exit(1) }
	// TODO: fmt.Println("token.TokenType:", tok.TokenType)
	// TODO: fmt.Println("token.Expiry:", tok.Expiry.Format(time.RFC3339))
	// TODO: fmt.Println("token.AccessToken[:12]:", tok.AccessToken[:12]+"...")

	_ = ctx
	_ = google.FindDefaultCredentials
	_ = os.Exit
	_ = fmt.Println
}
