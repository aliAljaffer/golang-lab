package s3util

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// fakeS3 records every call and returns canned responses. Same shape as the
// fakeNotifier in 06-testing/04-mock-interface — struct + slice, no framework.
type fakeS3 struct {
	calls []s3.GetObjectInput
	body  []byte
	err   error
}

func (f *fakeS3) GetObject(_ context.Context, in *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	f.calls = append(f.calls, *in)
	if f.err != nil {
		return nil, f.err
	}
	return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(f.body))}, nil
}

func TestFetchKey_ReturnsBody(t *testing.T) {
	f := &fakeS3{body: []byte("hello world")}

	got, err := FetchKey(context.Background(), f, "b", "k", 1024)
	if err != nil {
		t.Fatalf("FetchKey: %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("body = %q, want %q", got, "hello world")
	}
	if len(f.calls) != 1 || *f.calls[0].Bucket != "b" || *f.calls[0].Key != "k" {
		t.Errorf("calls = %+v, want one call with b/k", f.calls)
	}
}

func TestFetchKey_TruncatesAtMaxBytes(t *testing.T) {
	f := &fakeS3{body: bytes.Repeat([]byte("x"), 1000)}

	got, err := FetchKey(context.Background(), f, "b", "k", 100)
	if err != nil {
		t.Fatalf("FetchKey: %v", err)
	}
	if len(got) != 100 {
		t.Errorf("len(got) = %d, want 100 (LimitReader cap)", len(got))
	}
}

func TestFetchKey_PropagatesError(t *testing.T) {
	stub := errors.New("nosuchkey")
	f := &fakeS3{err: stub}

	_, err := FetchKey(context.Background(), f, "b", "k", 1024)
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}

func TestFetchKey_RejectsNonPositiveMax(t *testing.T) {
	f := &fakeS3{}

	if _, err := FetchKey(context.Background(), f, "b", "k", 0); err == nil {
		t.Error("FetchKey(maxBytes=0): err = nil, want validation error")
	}
	if len(f.calls) != 0 {
		t.Errorf("api was called %d time(s), want 0 (validation should short-circuit)", len(f.calls))
	}
}

// TODO: add a test for an `S3PutObjectAPI` interface (a second method) — the
// pattern composes: one interface per slice of the API surface a function uses.
