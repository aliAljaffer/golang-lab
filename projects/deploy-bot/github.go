package main

import (
	"context"
	"errors"
	"net/http"
)

// ErrReleaseNotFound is returned by ReleaseFetcher when the GitHub API
// returns 404 for the requested owner/repo/tag combination. Pipeline
// callers use errors.Is to detect this and exit with a clean message
// instead of a 5xx-looking stack trace.
var ErrReleaseNotFound = errors.New("release not found")

// ErrNoSuitableAsset is returned when a release exists but none of its
// attached assets match the selection rule (Dockerfile or *.tar.gz).
var ErrNoSuitableAsset = errors.New("no suitable asset on release")

// ReleaseFetcher locates the artifact URL for a (owner, repo, tag) tuple.
// Production impl is *GHReleaseFetcher (real GitHub API); tests use a fake
// returning programmed (url, err) values.
type ReleaseFetcher interface {
	Fetch(ctx context.Context, owner, repo, tag string) (artifactURL string, err error)
}

// asset is the JSON shape of a single item under release.assets in the
// GitHub API response. Only the fields we actually use are decoded.
type asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// release is the JSON shape of the GET /releases/tags/{tag} response.
type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

// GHReleaseFetcher calls the GitHub Releases API and picks one asset.
//
// Selection rule (matches PLAN.md):
//   - prefer an asset whose Name is exactly "Dockerfile"
//   - else the first asset whose Name has suffix ".tar.gz" (or ".tgz")
//   - else ErrNoSuitableAsset
//
// BaseURL is configurable so tests can point at an httptest.Server.
// Token, if non-empty, is sent as `Authorization: Bearer <token>`.
type GHReleaseFetcher struct {
	HTTPClient *http.Client
	Token      string
	BaseURL    string // e.g. "https://api.github.com"
}

// Fetch returns the URL of the chosen asset for owner/repo at tag.
//
// Wire contract:
//
//	GET {BaseURL}/repos/{owner}/{repo}/releases/tags/{tag}
//	Headers: Authorization: Bearer <token>   (if Token != "")
//	         Accept: application/vnd.github+json
//	Response: 200 -> JSON release object
//	          404 -> ErrReleaseNotFound
//	          other non-2xx -> error (include status in the message)
func (f *GHReleaseFetcher) Fetch(ctx context.Context, owner, repo, tag string) (string, error) {
	// TODO: hit the wire contract above. Decisions the tests pin:
	//   - the URL path shape is exact: /repos/{owner}/{repo}/releases/tags/{tag}.
	//   - 404 must return ErrReleaseNotFound (sentinel — tests use errors.Is).
	//   - other non-2xx must return a real error mentioning the status.
	//   - the Authorization header is only set when Token is non-empty;
	//     the unauthenticated path is a tested branch (public repos).
	//   - decode into the local `release` struct, then defer to pickAsset.
	return "", errors.New("GHReleaseFetcher.Fetch not implemented")
}

// pickAsset applies the selection rule on the assets attached to a release.
// Returns ErrNoSuitableAsset if nothing matches.
//
// Selection rule:
//  1. Exact-name match "Dockerfile" wins outright.
//  2. Otherwise, the first asset whose Name ends in ".tar.gz" or ".tgz".
//  3. Otherwise, ErrNoSuitableAsset.
//
// Pure; trivially unit-testable; called from Fetch.
func pickAsset(assets []asset) (string, error) {
	// TODO: apply the selection rule above. The "Dockerfile beats *.tar.gz
	//   even if the tarball appears first" priority is the part the test
	//   pins — one pass with a `fallback` variable is the simplest way.
	return "", ErrNoSuitableAsset
}
