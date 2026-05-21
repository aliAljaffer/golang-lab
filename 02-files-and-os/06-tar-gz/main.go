// 06-tar-gz — create a .tar.gz containing two small in-memory files,
// then read it back and print each entry's name + size.
//
// Run:
//   go run .
package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

type entry struct {
	name string
	data []byte
}

func main() {
	const path = "/tmp/go-learning-06.tar.gz"
	defer os.Remove(path)

	entries := []entry{
		{name: "hello.txt", data: []byte("hello from tar\n")},
		{name: "numbers.txt", data: []byte("1\n2\n3\n")},
	}

	// --- Write phase ---
	// TODO: os.Create(path); defer Close()
	// TODO: gzip.NewWriter(file); defer Close()
	// TODO: tar.NewWriter(gz);   defer Close()
	// TODO: for each entry: WriteHeader(&tar.Header{Name, Mode: 0o644, Size: int64(len(data))})
	//                       then Write(data)

	// --- Read phase ---
	// TODO: os.Open(path); defer Close()
	// TODO: gzip.NewReader(file); defer Close()
	// TODO: tar.NewReader(gz). Loop calling Next(); io.EOF means done.
	//       Print "<name> (<size> bytes)" for each header.

	_ = tar.Header{}
	_ = gzip.NewWriter
	_ = io.EOF
	_ = entries
	_ = fmt.Println
}
