package workspace

import (
	"fmt"
	"io"
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
	src := filepath.Join(sc.SourceDir(), "fixtures")
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("fixtures %s: %w", src, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("fixtures %s is not a directory", src)
	}
	return copyTree(src, filepath.Join(dest, "fixtures"))
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	_, err = io.Copy(out, in)
	return err
}
