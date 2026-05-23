package main

import (
	"context"
	"errors"
	"io"
	"net/http"
)

// Downloader fetches the artifact bytes for an asset URL. Returns an open
// ReadCloser the caller is responsible for closing.
//
// Production impl is *HTTPDownloader; tests use a fake returning a
// programmed body / error.
type Downloader interface {
	Download(ctx context.Context, url string) (io.ReadCloser, error)
}

// HTTPDownloader follows GitHub's 302 redirect from the API-host download
// URL to the S3-presigned URL.
//
// CRITICAL: when the redirect target is an S3-presigned URL, the request
// MUST NOT carry the GitHub `Authorization` header — S3 rejects requests
// that double-sign. Strip Authorization on cross-host redirects.
//
// Implement this via *http.Client.CheckRedirect: inspect req.URL.Host vs
// via[0].URL.Host and clear req.Header["Authorization"] when they differ.
type HTTPDownloader struct {
	Client *http.Client
	Token  string
}

// Download issues GET url and returns the body.
//
// Wire contract:
//
//	GET url
//	Headers: Authorization: Bearer <Token>   (if Token != "")
//	         Accept: application/octet-stream
//	Response: 2xx -> return resp.Body (caller closes)
//	          non-2xx -> error containing the status
func (d *HTTPDownloader) Download(ctx context.Context, url string) (io.ReadCloser, error) {
	// TODO: build a request with method GET, ctx, url.
	// TODO: set "Accept: application/octet-stream".
	// TODO: if d.Token != "" { req.Header.Set("Authorization", "Bearer " + d.Token) }.
	// TODO: pick the client. If d.Client == nil, build a fresh one with a
	// TODO: CheckRedirect that strips Authorization on cross-host redirects.
	// TODO: do the request; on non-2xx, drain+close body and return an error.
	// TODO: on success, return resp.Body, nil (caller closes).
	return nil, errors.New("HTTPDownloader.Download not implemented")
}

// stripAuthOnHostChange is the CheckRedirect helper Download uses to defuse
// the "GitHub token leaks into the S3 presigned URL" bug.
//
// Signature matches *http.Client.CheckRedirect: returning a non-nil error
// stops the redirect chain; returning nil follows it.
func stripAuthOnHostChange(req *http.Request, via []*http.Request) error {
	// TODO: if len(via) == 0 { return nil }.
	// TODO: prev := via[len(via)-1].
	// TODO: if req.URL.Host != prev.URL.Host { req.Header.Del("Authorization") }.
	// TODO: if len(via) >= 10 { return errors.New("too many redirects") }.
	// TODO: return nil.
	return nil
}
