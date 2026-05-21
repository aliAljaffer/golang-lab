//go:build exercise

package healthcheck

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newClient() *http.Client {
	return &http.Client{Timeout: 2 * time.Second}
}

func TestCheckAll_MixedStatuses(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		// 302 is non-2xx after redirect handling; but http.Client follows redirects by default,
		// so we serve a 404 at the destination to assert OK=false survives the chain.
		http.Redirect(w, r, "/notfound", http.StatusFound)
	})
	mux.HandleFunc("/notfound", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(404) })
	mux.HandleFunc("/boom", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) })
	srv := httptest.NewServer(mux)
	defer srv.Close()

	urls := []string{
		srv.URL + "/ok",
		srv.URL + "/boom",
		srv.URL + "/redirect",
	}
	got := CheckAll(newClient(), urls)
	if len(got) != 3 {
		t.Fatalf("len=%d, want 3", len(got))
	}

	cases := []struct {
		idx        int
		wantStatus int
		wantOK     bool
	}{
		{0, 200, true},
		{1, 500, false},
		{2, 404, false},
	}
	for _, c := range cases {
		if got[c.idx].URL != urls[c.idx] {
			t.Errorf("idx %d: URL = %q, want %q", c.idx, got[c.idx].URL, urls[c.idx])
		}
		if got[c.idx].Status != c.wantStatus {
			t.Errorf("idx %d: Status = %d, want %d", c.idx, got[c.idx].Status, c.wantStatus)
		}
		if got[c.idx].OK != c.wantOK {
			t.Errorf("idx %d: OK = %v, want %v", c.idx, got[c.idx].OK, c.wantOK)
		}
		if got[c.idx].Err != nil {
			t.Errorf("idx %d: Err = %v, want nil", c.idx, got[c.idx].Err)
		}
		if got[c.idx].Duration <= 0 {
			t.Errorf("idx %d: Duration = %v, want > 0", c.idx, got[c.idx].Duration)
		}
	}
}

func TestCheckAll_TransportErrorDoesNotAbort(t *testing.T) {
	// Spin up a server then close it to guarantee connection refused.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	deadURL := srv.URL
	srv.Close()

	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv2.Close()

	urls := []string{deadURL, srv2.URL}
	got := CheckAll(newClient(), urls)
	if len(got) != 2 {
		t.Fatalf("len=%d, want 2", len(got))
	}
	if got[0].Err == nil {
		t.Errorf("expected Err on dead URL, got nil (status=%d)", got[0].Status)
	}
	if got[0].OK {
		t.Errorf("dead URL should not be OK")
	}
	if got[1].Status != 200 || !got[1].OK {
		t.Errorf("live URL should be 200 OK, got %+v", got[1])
	}
}

func TestCheckAll_PreservesInputOrder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Each path returns a status equal to the numeric tail.
		switch r.URL.Path {
		case "/a":
			w.WriteHeader(201)
		case "/b":
			w.WriteHeader(202)
		case "/c":
			w.WriteHeader(204)
		}
	}))
	defer srv.Close()

	urls := []string{srv.URL + "/a", srv.URL + "/b", srv.URL + "/c"}
	got := CheckAll(newClient(), urls)
	if len(got) != 3 {
		t.Fatalf("len=%d, want 3", len(got))
	}
	wantStatuses := []int{201, 202, 204}
	for i, w := range wantStatuses {
		if got[i].URL != urls[i] {
			t.Errorf("i=%d URL = %q, want %q (order should be preserved)", i, got[i].URL, urls[i])
		}
		if got[i].Status != w {
			t.Errorf("i=%d Status = %d, want %d", i, got[i].Status, w)
		}
	}
}

func TestCheckAll_EmptyInput(t *testing.T) {
	got := CheckAll(newClient(), nil)
	if len(got) != 0 {
		t.Errorf("empty input should return empty result, got %+v", got)
	}
}
