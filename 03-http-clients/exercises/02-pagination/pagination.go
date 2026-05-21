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
	// TODO: split on "," (each entry is one link relation).
	// TODO: for each entry, strings.SplitN on ";" — first part is "<url>", rest are params.
	// TODO: find the entry whose params contain rel="next" — strip "<" and ">" from the url.
	// TODO: return that url, else "".
	return ""
}

// FetchAll calls GET on startURL, then follows `Link: rel="next"` to the
// next page, and so on until there are no more pages. Returns the page
// bodies in order. Stops and returns an error on any non-2xx response or
// transport error.
func FetchAll(client *http.Client, startURL string) ([][]byte, error) {
	// TODO: loop:
	//         resp, err := client.Get(url)
	//         if err -> return what we have + err
	//         if status != 2xx -> resp.Body.Close(); return err
	//         body, _ := io.ReadAll(resp.Body); resp.Body.Close()
	//         pages = append(pages, body)
	//         next := ParseNextLink(resp.Header.Get("Link"))
	//         if next == "" { break }
	//         url = next
	// TODO: return pages, nil.
	_ = startURL
	_ = client
	return nil, nil
}
