package notify

import (
	"errors"
	"testing"
)

// fakeNotifier is a hand-rolled mock. Records every call so the test can
// assert what was sent. No framework needed — a struct + a slice does the job.
type fakeNotifier struct {
	sent    []sentMessage
	failOn  string // if non-empty, Send returns errStub when `to` matches
}

type sentMessage struct {
	to, body string
}

var errStub = errors.New("delivery failed")

func (f *fakeNotifier) Send(to, body string) error {
	if f.failOn != "" && to == f.failOn {
		return errStub
	}
	f.sent = append(f.sent, sentMessage{to, body})
	return nil
}

func TestWelcome_DeliversGreeting(t *testing.T) {
	f := &fakeNotifier{}

	if err := Welcome(f, "Ali"); err != nil {
		t.Fatalf("Welcome: %v", err)
	}
	if len(f.sent) != 1 {
		t.Fatalf("len(sent) = %d, want 1", len(f.sent))
	}
	if got := f.sent[0]; got.to != "Ali" || got.body != "Welcome, Ali!" {
		t.Errorf("sent = %+v, want {to: Ali, body: Welcome, Ali!}", got)
	}
}

func TestWelcome_EmptyNameDoesNotCallNotifier(t *testing.T) {
	f := &fakeNotifier{}

	err := Welcome(f, "")
	if err == nil {
		t.Fatal("Welcome(\"\"): err = nil, want validation error")
	}
	if len(f.sent) != 0 {
		t.Errorf("notifier was called %d times, want 0 (validation should short-circuit)", len(f.sent))
	}
}

func TestWelcome_PropagatesNotifierError(t *testing.T) {
	f := &fakeNotifier{failOn: "Ali"}

	err := Welcome(f, "Ali")
	if !errors.Is(err, errStub) {
		t.Errorf("err = %v, want %v", err, errStub)
	}
}

// TODO: add a test that asserts Welcome NEVER sends an empty body — even if name is "  " (spaces). Then fix Welcome to trim/reject.
