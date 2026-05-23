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
			// TODO: pull credentials with r.BasicAuth, then compare each
			//   field with subtle.ConstantTimeCompare. Critical: compute
			//   BOTH comparisons (don't short-circuit with &&) so the timing
			//   doesn't leak which field was wrong. AND the two results
			//   together at the end.

			_ = subtle.ConstantTimeCompare
			unauthorized(w)
		})
	}
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}
