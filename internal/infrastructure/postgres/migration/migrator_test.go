package migration

import (
	"path/filepath"
	"testing"
)

func TestVersionFromFile(t *testing.T) {
	got := versionFromFile(filepath.Join("migrations", "000001_init.up.sql"))
	if got != "000001_init" {
		t.Fatalf("expected version 000001_init, got %s", got)
	}
}
