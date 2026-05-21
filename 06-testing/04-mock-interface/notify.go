// 04-mock-interface — define small interfaces at the consumption site, then
// inject a fake in tests. No mocking framework. No generated code. Plain Go.
package notify

import "fmt"

// Notifier is the seam: anyone who knows how to deliver a message qualifies.
// Real production code uses an SMTP/Slack/SNS implementation; tests use a
// pure-Go fake (see notify_test.go).
//
// Convention: define the interface where it is USED, not where it is satisfied.
// That way the consumer owns the minimal API surface it needs, and impls don't
// have to import a shared "interfaces" package.
type Notifier interface {
	Send(to, body string) error
}

// Welcome composes a greeting and asks the notifier to deliver it.
// Returns the error from the notifier so callers can react.
func Welcome(n Notifier, name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	return n.Send(name, fmt.Sprintf("Welcome, %s!", name))
}
