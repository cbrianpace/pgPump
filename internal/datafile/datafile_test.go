package datafile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetFilesInDir(t *testing.T) {
	dir := t.TempDir()

	files := map[string]bool{
		"DATA_orders.bin":  true,  // match
		"DATA_orders.csv":  true,  // match
		"DATA_orders.txt":  false, // unsupported extension
		"other_orders.bin": false, // wrong prefix
	}
	for name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	// A subdirectory that matches the filter must be ignored.
	if err := os.Mkdir(filepath.Join(dir, "DATA_sub.bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := GetFilesInDir(dir, "DATA_")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := map[string]bool{}
	for _, f := range got {
		found[f] = true
	}
	for name, want := range files {
		if found[name] != want {
			t.Errorf("file %q: included=%v, want %v", name, found[name], want)
		}
	}
	if found["DATA_sub.bin"] {
		t.Errorf("directory DATA_sub.bin should not be included")
	}
}

func TestGetFilesInDirMissing(t *testing.T) {
	if _, err := GetFilesInDir(filepath.Join(t.TempDir(), "nope"), "DATA_"); err == nil {
		t.Fatal("expected error for missing directory")
	}
}

func TestGetFilesInDirTableFilter(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"DATA_orders.bin", "DATA_customers.bin"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	got, err := GetFilesInDir(dir, "DATA_orders.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "DATA_orders.bin" {
		t.Errorf("got %v, want [DATA_orders.bin]", got)
	}
}
