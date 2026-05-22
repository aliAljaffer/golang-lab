//go:build exercise

package buildtar

import (
	"archive/tar"
	"bytes"
	"io"
	"testing"
)

// readBack untars b and returns the entries in order.
func readBack(t *testing.T, b []byte) []File {
	t.Helper()
	tr := tar.NewReader(bytes.NewReader(b))
	var out []File
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return out
		}
		if err != nil {
			t.Fatalf("tar.Next: %v", err)
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		out = append(out, File{Name: h.Name, Body: body, Mode: h.Mode})
	}
}

func TestBuildContext_SingleFile(t *testing.T) {
	in := []File{{Name: "Dockerfile", Body: []byte("FROM alpine:3\n"), Mode: 0644}}
	tarBytes, err := BuildContext(in)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}
	got := readBack(t, tarBytes)
	if len(got) != 1 {
		t.Fatalf("got %d entries, want 1", len(got))
	}
	if got[0].Name != "Dockerfile" || string(got[0].Body) != "FROM alpine:3\n" {
		t.Errorf("entry = %+v, want Dockerfile / FROM alpine:3", got[0])
	}
	if got[0].Mode != 0644 {
		t.Errorf("Mode = %o, want 0644", got[0].Mode)
	}
}

func TestBuildContext_MultipleFilesPreserveOrder(t *testing.T) {
	in := []File{
		{Name: "Dockerfile", Body: []byte("FROM scratch\nCOPY app /app\n"), Mode: 0644},
		{Name: "app", Body: []byte("\x7fELF..."), Mode: 0755},
		{Name: "readme.txt", Body: []byte("hi"), Mode: 0644},
	}
	tarBytes, err := BuildContext(in)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}
	got := readBack(t, tarBytes)
	if len(got) != 3 {
		t.Fatalf("got %d entries, want 3", len(got))
	}
	for i := range in {
		if got[i].Name != in[i].Name {
			t.Errorf("entry %d name = %q, want %q (order must be preserved)", i, got[i].Name, in[i].Name)
		}
		if !bytes.Equal(got[i].Body, in[i].Body) {
			t.Errorf("entry %d body mismatch", i)
		}
		if got[i].Mode != in[i].Mode {
			t.Errorf("entry %d Mode = %o, want %o", i, got[i].Mode, in[i].Mode)
		}
	}
}

func TestBuildContext_EmptyInputProducesValidEmptyTar(t *testing.T) {
	tarBytes, err := BuildContext(nil)
	if err != nil {
		t.Fatalf("BuildContext(nil): %v", err)
	}
	got := readBack(t, tarBytes) // must not panic; must terminate cleanly
	if len(got) != 0 {
		t.Errorf("got %d entries, want 0", len(got))
	}
}

func TestBuildContext_BinaryDataSurvives(t *testing.T) {
	// Build a payload with every byte value 0..255 to make sure tar headers
	// and length math are right even for non-textual data.
	body := make([]byte, 256)
	for i := range body {
		body[i] = byte(i)
	}
	in := []File{{Name: "blob", Body: body, Mode: 0644}}
	tarBytes, err := BuildContext(in)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}
	got := readBack(t, tarBytes)
	if len(got) != 1 {
		t.Fatalf("got %d entries, want 1", len(got))
	}
	if !bytes.Equal(got[0].Body, body) {
		t.Error("binary body did not round-trip")
	}
}

func TestBuildContext_TarIsClosedCleanly(t *testing.T) {
	// A tar without its terminator block is INVALID — tar.NewReader will
	// either error or hang. This catches "forgot to call tw.Close()".
	in := []File{{Name: "x", Body: []byte("y"), Mode: 0644}}
	tarBytes, err := BuildContext(in)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}
	tr := tar.NewReader(bytes.NewReader(tarBytes))
	if _, err := tr.Next(); err != nil {
		t.Fatalf("Next on first entry: %v", err)
	}
	if _, err := io.ReadAll(tr); err != nil {
		t.Fatalf("ReadAll body: %v", err)
	}
	// Next call MUST return io.EOF, not an error. If you forgot tw.Close(),
	// you'll get something like "unexpected EOF" instead.
	if _, err := tr.Next(); err != io.EOF {
		t.Errorf("second Next = %v, want io.EOF (tw.Close() missing?)", err)
	}
}
