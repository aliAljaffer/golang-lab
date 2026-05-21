// 01-greplite — a minimal grep.
//
// You implement: Grep(input io.Reader, pattern string, opts Options) ([]Match, error)
// The tests in greplite_test.go drive the design.
package greplite

import "io"

// Options controls how matching behaves.
type Options struct {
	IgnoreCase  bool // -i: case-insensitive
	LineNumbers bool // -n: include 1-based line number in Match
}

// Match describes one hit.
type Match struct {
	LineNumber int    // 1-based; 0 if Options.LineNumbers was false
	Text       string // the matching line, with the trailing newline stripped
}

// Grep scans input line by line and returns lines containing pattern.
// Empty pattern matches every line.
// Read with bufio.Scanner.
func Grep(input io.Reader, pattern string, opts Options) ([]Match, error) {
	// TODO: implement.
	return nil, nil
}
