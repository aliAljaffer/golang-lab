// Package untagged finds EC2 instances that are missing a required tag.
// Classic compliance check: every instance must have an Owner / CostCenter /
// Env tag, anything without it is reported.
//
// Exercise surface:
//
//	type EC2API interface { DescribeInstances paginatable }
//	func FindUntagged(ctx, api, requiredKey string) ([]string, error)
//
// Returns instance IDs in the order they came back from EC2.
package untagged

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// EC2API is the slice of *ec2.Client this package uses.
type EC2API interface {
	DescribeInstances(ctx context.Context, in *ec2.DescribeInstancesInput, opts ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

// FindUntagged returns the IDs of every instance that does NOT have a tag
// with Key == requiredKey. State doesn't matter (a "stopped" untagged instance
// still counts — it'll be there when someone starts it again).
//
// Hints:
//   - DescribeInstances output has .Reservations, each with .Instances.
//   - Each instance has .InstanceId (*string) and .Tags ([]ec2types.Tag).
//   - Each tag has .Key (*string) and .Value (*string).
//   - Use the paginator: ec2.NewDescribeInstancesPaginator(api, in).
//   - If two reservations are returned, both flatten into the output.
func FindUntagged(ctx context.Context, api EC2API, requiredKey string) ([]string, error) {
	if requiredKey == "" {
		return nil, errors.New("requiredKey must not be empty")
	}
	// TODO: implement.
	return nil, errors.New("FindUntagged not implemented")
}
