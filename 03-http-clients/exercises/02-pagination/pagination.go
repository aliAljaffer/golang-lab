// 02-pagination — paginate through GitHub-style endpoints that return
// a `Link: <next-url>; rel="next"` header.
//
// You implement two functions:
//   ParseNextLink(linkHeader string) string
//   FetchAll(client *http.Client, startURL string) ([][]byte, error)
//
// The tests in pagination_test.go drive the design.
package pagination

import "net/http"

// ParseNextLink extracts the `rel="next"` URL from an RFC-5988 Link header.
// Returns "" if there is no `rel="next"` link (i.e. we're on the last page).
//
// Example input (all on one line):
//   <https://api.github.com/repos/golang/go/issues?page=2>; rel="next",
//   <https://api.github.com/repos/golang/go/issues?page=10>; rel="last"
//
// Should return: https://api.github.com/repos/golang/go/issues?page=2
//
// Empty input → "" with no error.
func ParseNextLink(linkHeader string) string {
	// TODO: parse the RFC-5988 Link header. Each comma-separated entry has
	//   the shape `<url>; rel="something"`; you want the url whose rel is
	//   "next". Empty header is the "last page" signal — return "".
	return ""
}

// FetchAll calls GET on startURL, then follows `Link: rel="next"` to the
// next page, and so on until there are no more pages. Returns the page
// bodies in order. Stops and returns an error on any non-2xx response or
// transport error.
func FetchAll(client *http.Client, startURL string) ([][]byte, error) {
	// TODO: walk the pagination chain, appending each body. Two things to
	//   watch out for:
	//     - close every resp.Body, even on the non-2xx error path, or you
	//       leak the connection.
	//     - terminate on empty Link header (last page) — not on a specific
	//       status code. Some APIs page until the body is empty; GitHub
	//       pages until the header drops.
	_ = startURL
	_ = client
	return nil, nil
}
