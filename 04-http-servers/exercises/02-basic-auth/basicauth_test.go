//go:build exercise

package basicauth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func protectedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("secret"))
	})
}

func newRequest(t *testing.T, srvURL, user, pass string, setAuth bool) *http.Response {
	t.Helper()
	req, err := http.NewRequest("GET", srvURL+"/", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if setAuth {
		req.SetBasicAuth(user, pass)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	return resp
}

func TestBasicAuth_ValidCreds(t *testing.T) {
	h := BasicAuth("admin", "swordfish")(protectedHandler())
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp := newRequest(t, srv.URL, "admin", "swordfish", true)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestBasicAuth_NoHeader(t *testing.T) {
	h := BasicAuth("admin", "swordfish")(protectedHandler())
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp := newRequest(t, srv.URL, "", "", false)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
	if got := resp.Header.Get("WWW-Authenticate"); got == "" {
		t.Error("missing WWW-Authenticate header on 401")
	}
}

func TestBasicAuth_WrongPassword(t *testing.T) {
	h := BasicAuth("admin", "swordfish")(protectedHandler())
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp := newRequest(t, srv.URL, "admin", "wrong", true)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestBasicAuth_WrongUsername(t *testing.T) {
	h := BasicAuth("admin", "swordfish")(protectedHandler())
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp := newRequest(t, srv.URL, "guest", "swordfish", true)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestBasicAuth_DoesNotLeakInnerHandlerOnFailure(t *testing.T) {
	var inner int
	h := BasicAuth("admin", "swordfish")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inner++
	}))
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp := newRequest(t, srv.URL, "admin", "wrong", true)
	resp.Body.Close()
	if inner != 0 {
		t.Errorf("inner handler called %d times on failed auth, want 0", inner)
	}
}
