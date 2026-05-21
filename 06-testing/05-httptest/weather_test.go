package weather

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestClient_Get_HappyPath spins up an in-process HTTP server, returns a
// canned JSON payload, and asserts the client decodes it.
//
// `srv.Client()` returns a pre-wired *http.Client that knows how to talk to
// `srv.URL`. For plain-HTTP servers this is the same as &http.Client{}, but
// for httptest.NewTLSServer it sets up cert trust. Always use srv.Client() —
// it removes a class of "works in tests but not in TLS" bugs.
func TestClient_Get_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/forecast" {
			t.Errorf("unexpected path: %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("city"); got != "Riyadh" {
			t.Errorf("city query = %q, want Riyadh", got)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"city":"Riyadh","temp_c":42.5}`)
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BaseURL: srv.URL}
	got, err := c.Get("Riyadh")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.City != "Riyadh" || got.TempC != 42.5 {
		t.Errorf("got = %+v", got)
	}
}

func TestClient_Get_Non200IsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), BaseURL: srv.URL}
	if _, err := c.Get("anywhere"); err == nil {
		t.Fatal("expected error on 500, got nil")
	}
}

// TODO: add a test where the handler returns malformed JSON. Assert the error mentions "decode".
// TODO: add a test that injects a slow handler (time.Sleep) and a client with a short Timeout — assert a timeout error.
