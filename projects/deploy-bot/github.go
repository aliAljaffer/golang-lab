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
	// TODO: build the URL: f.BaseURL + "/repos/" + owner + "/" + repo + "/releases/tags/" + tag.
	// TODO: build a *http.Request with method GET and ctx.
	// TODO: set "Accept: application/vnd.github+json".
	// TODO: if f.Token != "" { req.Header.Set("Authorization", "Bearer " + f.Token) }.
	// TODO: pick client := f.HTTPClient (or http.DefaultClient if nil); do the request.
	// TODO: defer resp.Body.Close().
	// TODO: switch resp.StatusCode {
	// TODO:   case 200: // ok
	// TODO:   case 404: return "", ErrReleaseNotFound
	// TODO:   default: return "", fmt.Errorf("github releases API: %s", resp.Status)
	// TODO: }
	// TODO: decode the response body into a `release` struct (json.NewDecoder).
	// TODO: call pickAsset(rel.Assets) to get the chosen URL; return it.
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
	// TODO: var fallback string.
	// TODO: for _, a := range assets {
	// TODO:     if a.Name == "Dockerfile" { return a.URL, nil }
	// TODO:     if fallback == "" && (hasSuffix(a.Name, ".tar.gz") || hasSuffix(a.Name, ".tgz")) { fallback = a.URL }
	// TODO: }
	// TODO: if fallback != "" { return fallback, nil }
	// TODO: return "", ErrNoSuitableAsset.
	return "", ErrNoSuitableAsset
}
