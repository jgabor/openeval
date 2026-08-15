package workspace

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jgabor/openeval/internal/scenario"
)

func RoundDir(runDir, taskID string, round int) string {
	return filepath.Join(runDir, "tasks", taskID, fmt.Sprintf("round-%d", round))
}

func Seed(sc scenario.Scenario, dest string) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	src, err := fs.Sub(sc.SourceFS(), "fixtures")
	if err != nil {
		return fmt.Errorf("fixtures: %w", err)
	}
	if _, err := fs.Stat(src, "."); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("fixtures: %w", err)
	}
	return copyFSTree(src, filepath.Join(dest, "fixtures"))
}

func copyFSTree(src fs.FS, dst string) error {
	return fs.WalkDir(src, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dst, filepath.FromSlash(path))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := src.Open(path)
		if err != nil {
			return err
		}
		copyErr := copyReader(in, target)
		_ = in.Close()
		return copyErr
	})
}

func copyReader(in io.Reader, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	_, err = io.Copy(out, in)
	return err
}
