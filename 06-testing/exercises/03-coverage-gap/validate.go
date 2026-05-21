// Package validate enforces config rules.
//
// This file has several branches. Your job is to write tests until
// `go test -cover` reports 100.0% coverage. Use `go tool cover -html` to
// see which lines are still uncovered.
package validate

import (
	"errors"
	"fmt"
	"strings"
)

// Config is what we validate.
type Config struct {
	Name    string
	Port    int
	Tags    []string
	Env     string // must be "dev" | "staging" | "prod"
	Timeout int    // seconds; 0 means unlimited
}

var ErrEmptyName = errors.New("name is required")

// Validate returns the first rule violation it finds, or nil.
func Validate(c Config) error {
	if c.Name == "" {
		return ErrEmptyName
	}
	if len(c.Name) > 64 {
		return fmt.Errorf("name too long: %d chars (max 64)", len(c.Name))
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port %d out of range [1, 65535]", c.Port)
	}
	if c.Port < 1024 && c.Env == "prod" {
		return fmt.Errorf("port %d is privileged; prod must use >= 1024", c.Port)
	}
	switch c.Env {
	case "dev", "staging", "prod":
	default:
		return fmt.Errorf("unknown env %q", c.Env)
	}
	for _, tag := range c.Tags {
		if strings.ContainsAny(tag, " \t\n") {
			return fmt.Errorf("tag %q contains whitespace", tag)
		}
	}
	if c.Timeout < 0 {
		return fmt.Errorf("timeout %d must be >= 0", c.Timeout)
	}
	return nil
}
