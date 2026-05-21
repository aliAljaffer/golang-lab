// 01-read-write — the four file-I/O building blocks.
//
// Goal: write a small file, read it back, print its contents, then delete it.
//
// Run:
//   go run .
package main

import (
	"fmt"
	"os"
)

func main() {
	const path = "/tmp/go-learning-01-read-write.txt"
	payload := []byte("hello from os.WriteFile\n")

	// TODO: os.WriteFile(path, payload, 0o644). Handle the error.

	// TODO: data, err := os.ReadFile(path). Handle the error. Print the contents to stdout.

	// TODO: os.Remove(path) to clean up. Handle the error.

	_ = payload
	_ = fmt.Println
	_ = os.WriteFile
}
