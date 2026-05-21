// 02-basic-auth — HTTP Basic Auth middleware with constant-time comparison.
//
// You implement:
//   BasicAuth(username, password string) func(http.Handler) http.Handler
//
// Behavior:
//   - If r.BasicAuth() returns a matching user+pass, call next.
//   - Otherwise reply 401 with `WWW-Authenticate: Basic realm="restricted"`.
//   - Compare with crypto/subtle.ConstantTimeCompare — never `==` or `bytes.Equal`.
//
// Why constant-time? `==` short-circuits on the first differing byte. An attacker
// who can measure response time can guess credentials byte by byte. Auth code
// must always take the same time regardless of where the mismatch is.
package basicauth

import (
	"crypto/subtle"
	"net/http"
)

func BasicAuth(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: u, p, ok := r.BasicAuth()
			// TODO: if !ok { -> unauthorized(w); return }
			// TODO: userOK := subtle.ConstantTimeCompare([]byte(u), []byte(username)) == 1
			//       passOK := subtle.ConstantTimeCompare([]byte(p), []byte(password)) == 1
			// TODO: if !(userOK && passOK) { unauthorized(w); return }
			// TODO: next.ServeHTTP(w, r)

			_ = subtle.ConstantTimeCompare
			unauthorized(w)
		})
	}
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}
