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
	// TODO: do the GET per the wire contract above. The load-bearing detail
	//   is that resp.Body is RETURNED to the caller (you don't close it on
	//   success). Non-2xx is your responsibility to drain + close before
	//   returning the error — otherwise the connection won't go back to the
	//   pool. If d.Client is nil, you'll need to wire up stripAuthOnHostChange
	//   yourself so the GitHub token doesn't leak to S3.
	return nil, errors.New("HTTPDownloader.Download not implemented")
}

// stripAuthOnHostChange is the CheckRedirect helper Download uses to defuse
// the "GitHub token leaks into the S3 presigned URL" bug.
//
// Signature matches *http.Client.CheckRedirect: returning a non-nil error
// stops the redirect chain; returning nil follows it.
func stripAuthOnHostChange(req *http.Request, via []*http.Request) error {
	// TODO: delete the Authorization header from req when the host differs
	//   from the previous hop in `via`. Also bail out if the chain gets
	//   silly long — net/http's default is 10, mirror that.
	return nil
}
