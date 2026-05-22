//go:build exercise

package fileinfods

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestReadFileInfo_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := ReadFileInfo(path)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !info.Exists {
		t.Fatal("Exists=false, want true")
	}
	if info.Size != 5 {
		t.Fatalf("Size=%d, want 5", info.Size)
	}
	if info.Path != path {
		t.Fatalf("Path=%q, want %q", info.Path, path)
	}
}

func TestReadFileInfo_MissingFileIsNotAnError(t *testing.T) {
	// "Does this file exist?" is a valid question. A missing file should
	// return Exists=false, NOT an error — otherwise data sources can't be
	// used in `count = data.x.exists ? 1 : 0` patterns.
	path := filepath.Join(t.TempDir(), "ghost.txt")
	info, err := ReadFileInfo(path)
	if err != nil {
		t.Fatalf("missing file should not error, got %v", err)
	}
	if info.Exists {
		t.Fatal("Exists=true for missing file")
	}
}

func TestReadFileInfo_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := ReadFileInfo(path)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !info.Exists {
		t.Fatal("Exists=false for empty file (should be true)")
	}
	if info.Size != 0 {
		t.Fatalf("Size=%d, want 0", info.Size)
	}
}

func TestMetadata_TypeNameUsesProviderPrefix(t *testing.T) {
	d := NewFileInfoDS()
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "fileops"}, &resp)
	if resp.TypeName != "fileops_file_info" {
		t.Fatalf("TypeName=%q, want fileops_file_info", resp.TypeName)
	}
}

func TestSchema_HasThreeAttributes(t *testing.T) {
	d := NewFileInfoDS()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	attrs := resp.Schema.Attributes
	want := []string{"path", "exists", "size"}
	for _, name := range want {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing %q attribute", name)
		}
	}
}
