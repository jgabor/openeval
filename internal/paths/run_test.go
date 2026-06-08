package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNextRunIndex(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "scenarios", "demo", "runs")
	n1, err := nextRunIndex(base)
	if err != nil || n1 != 1 {
		t.Fatalf("first index: got %d err %v", n1, err)
	}
	if err := os.MkdirAll(filepath.Join(base, "1"), 0o755); err != nil {
		t.Fatal(err)
	}
	n2, err := nextRunIndex(base)
	if err != nil || n2 != 2 {
		t.Fatalf("second index: got %d err %v", n2, err)
	}
}

func TestResolveRunDirVariationOverwrite(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	first, err := ResolveRunDir("example-fixtures", "baseline", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(first); err != nil {
		t.Fatal(err)
	}
	second, err := ResolveRunDir("example-fixtures", "baseline", "")
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("expected same variation dir, got %q and %q", first, second)
	}
}
