// Package unlabeled finds GCE instances missing a required label key.
// Compliance check: every VM should carry Owner / CostCenter / Env labels;
// anything without it is reported.
//
// Exercise surface:
//
//	type ComputeAPI interface { AggregatedList GCE instances across zones }
//	func FindUnlabeled(ctx, api, project, requiredKey) ([]string, error)
//
// Returns instance NAMES in the order ComputeAPI yielded them (zone order
// first, then instance order within zone). The GCP-flavoured cousin of
// 07-aws/exercises/02-find-untagged.
package unlabeled

import (
	"context"
	"errors"
)

// InstanceSummary is the minimal projection of *computepb.Instance needed here.
type InstanceSummary struct {
	Name   string
	Zone   string            // "zones/us-central1-a" (informational only)
	Labels map[string]string // GCE labels — server-modeled map, not a tag slice
}

// ComputeAPI is the slice of compute this package uses. Same wrapper-style
// pattern as 07-mocking-gcs — we don't try to interface
// *compute.InstancesClient directly because its `AggregatedList` returns a
// concrete `*InstanceIterator`.
type ComputeAPI interface {
	// AggregatedListInstances returns every instance across every zone in
	// <project>, already flattened. Iterator-of-zones is the wrapper's job;
	// callers see one flat slice.
	AggregatedListInstances(ctx context.Context, project string) ([]InstanceSummary, error)
}

// FindUnlabeled returns the names of every instance whose Labels map does
// NOT contain the requiredKey. State doesn't matter (a TERMINATED unlabeled
// instance still counts — it'll be back when someone starts it again).
//
// Hints:
//   - One call: instances, err := api.AggregatedListInstances(ctx, project)
//   - For each instance: check if instance.Labels[requiredKey] exists.
//     A non-empty value is required-key-PRESENT (the key existing is what
//     matters; empty values like Owner="" are valid).
//   - Return names in the order the API returned them.
//
// Note: GCP labels are case-insensitive on the server (keys are stored
// lowercased). The fake honours that — match against requiredKey as-is, but
// be aware production might surprise you with case folding.
func FindUnlabeled(ctx context.Context, api ComputeAPI, project, requiredKey string) ([]string, error) {
	if requiredKey == "" {
		return nil, errors.New("requiredKey must not be empty")
	}
	// TODO: implement.
	return nil, errors.New("FindUnlabeled not implemented")
}
