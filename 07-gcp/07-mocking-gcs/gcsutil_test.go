package gcsutil

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

// fakeGCS records every call and returns canned responses. Same hand-rolled
// pattern as 07-aws/07-mocking-sdk's fakeS3 — struct + slices, no framework.
type fakeGCS struct {
	// recorded calls.
	readCalls  []readCall
	writeCalls []writeCall
	listCalls  []listCall

	// primed responses. Tests prime these; the fake returns them.
	readBody  []byte
	readErr   error
	writeErr  error
	listObjs  []ObjectInfo
	listErr   error
}

type readCall struct{ Bucket, Key string; MaxBytes int64 }
type writeCall struct{ Bucket, Key string; Body []byte }
type listCall struct{ Bucket, Prefix string }

func (f *fakeGCS) Read(_ context.Context, bucket, key string, maxBytes int64) ([]byte, error) {
	f.readCalls = append(f.readCalls, readCall{bucket, key, maxBytes})
	if f.readErr != nil {
		return nil, f.readErr
	}
	if int64(len(f.readBody)) > maxBytes {
		return f.readBody[:maxBytes], nil
	}
	return f.readBody, nil
}

func (f *fakeGCS) Write(_ context.Context, bucket, key string, body []byte) error {
	// Defensive copy — tests sometimes mutate the input slice after the call.
	cp := make([]byte, len(body))
	copy(cp, body)
	f.writeCalls = append(f.writeCalls, writeCall{bucket, key, cp})
	return f.writeErr
}

func (f *fakeGCS) List(_ context.Context, bucket, prefix string) ([]ObjectInfo, error) {
	f.listCalls = append(f.listCalls, listCall{bucket, prefix})
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listObjs, nil
}

// ---- FetchKey --------------------------------------------------------------

func TestFetchKey_ReturnsBody(t *testing.T) {
	f := &fakeGCS{readBody: []byte("hello gcs")}

	got, err := FetchKey(context.Background(), f, "b", "k", 1024)
	if err != nil {
		t.Fatalf("FetchKey: %v", err)
	}
	if string(got) != "hello gcs" {
		t.Errorf("body = %q, want %q", got, "hello gcs")
	}
	if len(f.readCalls) != 1 {
		t.Fatalf("Read calls = %d, want 1", len(f.readCalls))
	}
	if c := f.readCalls[0]; c.Bucket != "b" || c.Key != "k" || c.MaxBytes != 1024 {
		t.Errorf("call = %+v, want {b k 1024}", c)
	}
}

func TestFetchKey_TruncatesAtMaxBytes(t *testing.T) {
	f := &fakeGCS{readBody: bytes.Repeat([]byte("x"), 1000)}

	got, err := FetchKey(context.Background(), f, "b", "k", 100)
	if err != nil {
		t.Fatalf("FetchKey: %v", err)
	}
	if len(got) != 100 {
		t.Errorf("len(got) = %d, want 100 (Read should cap at maxBytes)", len(got))
	}
}

func TestFetchKey_PropagatesError(t *testing.T) {
	stub := errors.New("object not found")
	f := &fakeGCS{readErr: stub}

	_, err := FetchKey(context.Background(), f, "b", "k", 1024)
	if !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}

func TestFetchKey_RejectsNonPositiveMax(t *testing.T) {
	f := &fakeGCS{}

	if _, err := FetchKey(context.Background(), f, "b", "k", 0); err == nil {
		t.Error("FetchKey(maxBytes=0): err = nil, want validation error")
	}
	if len(f.readCalls) != 0 {
		t.Errorf("api was called %d time(s), want 0 (validation should short-circuit)", len(f.readCalls))
	}
}

// ---- TotalSize -------------------------------------------------------------

func TestTotalSize_SumsAllObjects(t *testing.T) {
	when := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	f := &fakeGCS{listObjs: []ObjectInfo{
		{Name: "a", Size: 100, Updated: when},
		{Name: "b", Size: 250, Updated: when},
		{Name: "c", Size: 7, Updated: when},
	}}

	got, err := TotalSize(context.Background(), f, "b", "logs/")
	if err != nil {
		t.Fatalf("TotalSize: %v", err)
	}
	if got != 357 {
		t.Errorf("total = %d, want 357", got)
	}
	if len(f.listCalls) != 1 || f.listCalls[0].Prefix != "logs/" {
		t.Errorf("list calls = %+v, want one call with prefix=logs/", f.listCalls)
	}
}

func TestTotalSize_EmptyListReturnsZero(t *testing.T) {
	f := &fakeGCS{}

	got, err := TotalSize(context.Background(), f, "b", "")
	if err != nil {
		t.Fatalf("TotalSize: %v", err)
	}
	if got != 0 {
		t.Errorf("total = %d, want 0", got)
	}
}

func TestTotalSize_PropagatesListError(t *testing.T) {
	stub := errors.New("permission denied")
	f := &fakeGCS{listErr: stub}

	if _, err := TotalSize(context.Background(), f, "b", ""); !errors.Is(err, stub) {
		t.Errorf("err = %v, want %v", err, stub)
	}
}
