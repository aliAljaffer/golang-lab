//go:build exercise

package untagged

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// fakeEC2 returns canned reservations.
type fakeEC2 struct {
	pages [][]ec2types.Reservation // each element is one page
	idx   int
	err   error
}

func (f *fakeEC2) DescribeInstances(_ context.Context, _ *ec2.DescribeInstancesInput, _ ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.idx >= len(f.pages) {
		return &ec2.DescribeInstancesOutput{}, nil
	}
	out := &ec2.DescribeInstancesOutput{Reservations: f.pages[f.idx]}
	f.idx++
	if f.idx < len(f.pages) {
		next := "next-token"
		out.NextToken = &next
	}
	return out, nil
}

func ptr(s string) *string { return &s }

func inst(id string, tags ...string) ec2types.Instance {
	in := ec2types.Instance{InstanceId: ptr(id)}
	for i := 0; i+1 < len(tags); i += 2 {
		k, v := tags[i], tags[i+1]
		in.Tags = append(in.Tags, ec2types.Tag{Key: ptr(k), Value: ptr(v)})
	}
	return in
}

func TestFindUntagged_FlagsMissingKey(t *testing.T) {
	f := &fakeEC2{pages: [][]ec2types.Reservation{{
		{Instances: []ec2types.Instance{
			inst("i-good", "Owner", "ali"),
			inst("i-bad"), // no tags
		}},
	}}}

	got, err := FindUntagged(context.Background(), f, "Owner")
	if err != nil {
		t.Fatalf("FindUntagged: %v", err)
	}
	if len(got) != 1 || got[0] != "i-bad" {
		t.Errorf("got %v, want [i-bad]", got)
	}
}

func TestFindUntagged_EmptyValueIsStillPresent(t *testing.T) {
	// Per AWS semantics, the tag KEY existing is what counts. Empty values
	// are valid (and common — "Owner=" might be set to "unknown" later).
	f := &fakeEC2{pages: [][]ec2types.Reservation{{
		{Instances: []ec2types.Instance{inst("i-empty", "Owner", "")}},
	}}}

	got, err := FindUntagged(context.Background(), f, "Owner")
	if err != nil {
		t.Fatalf("FindUntagged: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %v, want [] (key exists, value is empty but valid)", got)
	}
}

func TestFindUntagged_FlattensReservations(t *testing.T) {
	f := &fakeEC2{pages: [][]ec2types.Reservation{{
		{Instances: []ec2types.Instance{inst("i-1")}},
		{Instances: []ec2types.Instance{inst("i-2", "Owner", "x"), inst("i-3")}},
	}}}

	got, err := FindUntagged(context.Background(), f, "Owner")
	if err != nil {
		t.Fatalf("FindUntagged: %v", err)
	}
	if strings.Join(got, ",") != "i-1,i-3" {
		t.Errorf("got %v, want [i-1 i-3] (in order across both reservations)", got)
	}
}

func TestFindUntagged_PagesThroughResults(t *testing.T) {
	f := &fakeEC2{pages: [][]ec2types.Reservation{
		{{Instances: []ec2types.Instance{inst("i-p1-a"), inst("i-p1-b", "Owner", "x")}}},
		{{Instances: []ec2types.Instance{inst("i-p2-a", "Owner", "y"), inst("i-p2-b")}}},
	}}

	got, err := FindUntagged(context.Background(), f, "Owner")
	if err != nil {
		t.Fatalf("FindUntagged: %v", err)
	}
	if strings.Join(got, ",") != "i-p1-a,i-p2-b" {
		t.Errorf("got %v, want [i-p1-a i-p2-b] (paginator must consume both pages)", got)
	}
}

func TestFindUntagged_PropagatesError(t *testing.T) {
	stub := errors.New("rate exceeded")
	f := &fakeEC2{err: stub}

	_, err := FindUntagged(context.Background(), f, "Owner")
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}
