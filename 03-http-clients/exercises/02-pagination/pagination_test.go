//go:build exercise

package pagination

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newClient() *http.Client {
	return &http.Client{Timeout: 2 * time.Second}
}

func TestParseNextLink_Found(t *testing.T) {
	in := `<https://api.example.com/issues?page=2>; rel="next", <https://api.example.com/issues?page=10>; rel="last"`
	got := ParseNextLink(in)
	want := "https://api.example.com/issues?page=2"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseNextLink_NoNext(t *testing.T) {
	// Only "prev" and "last" — we're on the final page.
	in := `<https://api.example.com/issues?page=9>; rel="prev", <https://api.example.com/issues?page=10>; rel="last"`
	if got := ParseNextLink(in); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestParseNextLink_Empty(t *testing.T) {
	if got := ParseNextLink(""); got != "" {
		t.Errorf("empty input should give empty output, got %q", got)
	}
}

func TestFetchAll_FollowsLinkHeaderUntilDone(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/items":
			w.Header().Set("Link", fmt.Sprintf(`<%s/items?page=2>; rel="next"`, srv.URL))
			_, _ = fmt.Fprint(w, `["a","b"]`)
		case "/items?page=2":
		default:
			// Querystring isn't part of Path; check RawQuery for page 2/3.
		}
		// Use Path + RawQuery to disambiguate.
		q := r.URL.RawQuery
		if r.URL.Path == "/items" && q == "page=2" {
			w.Header().Set("Link", fmt.Sprintf(`<%s/items?page=3>; rel="next"`, srv.URL))
			_, _ = fmt.Fprint(w, `["c","d"]`)
			return
		}
		if r.URL.Path == "/items" && q == "page=3" {
			// No Link header -> end of pagination.
			_, _ = fmt.Fprint(w, `["e"]`)
			return
		}
	}))
	defer srv.Close()

	pages, err := FetchAll(newClient(), srv.URL+"/items")
	if err != nil {
		t.Fatalf("FetchAll: %v", err)
	}
	if len(pages) != 3 {
		t.Fatalf("got %d pages, want 3: %s", len(pages), pages)
	}
	wantBodies := []string{`["a","b"]`, `["c","d"]`, `["e"]`}
	for i, want := range wantBodies {
		if string(pages[i]) != want {
			t.Errorf("page %d = %q, want %q", i, pages[i], want)
		}
	}
}

func TestFetchAll_StopsOnError(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.RawQuery {
		case "":
			w.Header().Set("Link", fmt.Sprintf(`<%s/items?page=2>; rel="next"`, srv.URL))
			_, _ = fmt.Fprint(w, `["a"]`)
		case "page=2":
			w.WriteHeader(500)
		}
	}))
	defer srv.Close()

	pages, err := FetchAll(newClient(), srv.URL+"/items")
	if err == nil {
		t.Fatalf("expected error on 500, got nil (pages=%d)", len(pages))
	}
}

func TestFetchAll_SinglePageNoLink(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `["only"]`)
	}))
	defer srv.Close()

	pages, err := FetchAll(newClient(), srv.URL+"/items")
	if err != nil {
		t.Fatalf("FetchAll: %v", err)
	}
	if len(pages) != 1 || string(pages[0]) != `["only"]` {
		t.Errorf("got %v, want one page with [\"only\"]", pages)
	}
}
