// Package cleanup deletes GCS objects under a prefix older than a cutoff time.
// Classic janitor — every team has a `tmp/` prefix that nobody wants to
// hand-clean.
//
// Exercise surface:
//
//	type GCSAPI interface { ListObjects under prefix; DeleteObject }
//	func Cleanup(ctx, api, bucket, prefix, cutoff, dryRun) ([]string, error)
//
// Returns the names that were (or would have been) deleted. Order matches
// the iterator order. GCP-flavoured cousin of 07-aws/exercises/03-cleanup-old.
package cleanup

import (
	"context"
	"errors"
	"time"
)

// ObjectAttrs is the minimal projection of *storage.ObjectAttrs needed here.
type ObjectAttrs struct {
	Name    string
	Updated time.Time // GCS's last-modified timestamp
}

// GCSAPI is the slice of GCS this package uses. Two methods —
// list-under-prefix and per-key delete.
type GCSAPI interface {
	ListObjects(ctx context.Context, bucket, prefix string) ([]ObjectAttrs, error)
	DeleteObject(ctx context.Context, bucket, name string) error
}

// Cleanup deletes every object in <bucket> under <prefix> whose Updated time
// is BEFORE cutoff. If dryRun is true, no DeleteObject calls are made — the
// return list is what WOULD have been deleted.
//
// Hints:
//   - List once with api.ListObjects(ctx, bucket, prefix). The prefix narrows
//     server-side (no client-side filter needed for prefix).
//   - For each object: compare obj.Updated.Before(cutoff). If yes:
//     - If dryRun: append the name to the result, skip Delete.
//     - Else: call api.DeleteObject(ctx, bucket, obj.Name); append on success.
//   - If a DeleteObject errors, return the error AND the names deleted so far
//     (partial progress, so a retry can be informed).
//   - List errors abort with no names returned.
func Cleanup(ctx context.Context, api GCSAPI, bucket, prefix string, cutoff time.Time, dryRun bool) ([]string, error) {
	if bucket == "" {
		return nil, errors.New("bucket must not be empty")
	}
	// TODO: implement.
	return nil, errors.New("Cleanup not implemented")
}
