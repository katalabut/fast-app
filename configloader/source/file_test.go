package source

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileCandidates(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(file, []byte("port: 1234\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFile(file + "/child"); err == nil {
		t.Fatal("non-directory path component accepted")
	}
	source, err := NewFile("", dir, file+"/child", file)
	if err != nil {
		t.Fatal(err)
	}
	if source.path != file {
		t.Fatalf("selected %q, want %q", source.path, file)
	}
}
