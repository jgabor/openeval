package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ResolveRunDir(scenarioID, variation, override string) (string, error) {
	if override != "" {
		abs, err := filepath.Abs(override)
		if err != nil {
			return "", err
		}
		if err := os.MkdirAll(abs, 0o755); err != nil {
			return "", err
		}
		return abs, nil
	}
	base := filepath.Join("scenarios", scenarioID, "runs")
	if variation == "" || variation == "default" {
		n, err := nextRunIndex(base)
		if err != nil {
			return "", err
		}
		dir := filepath.Join(base, strconv.Itoa(n))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
		return dir, nil
	}
	dir := filepath.Join(base, variation)
	if err := os.RemoveAll(dir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func nextRunIndex(base string) (int, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}
	max := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		n, err := strconv.Atoi(e.Name())
		if err == nil && n > max {
			max = n
		}
	}
	return max + 1, nil
}

func ScorePath(runDir string) string {
	return filepath.Join(runDir, "score.json")
}

func ValidateCompareDirs(a, b string) error {
	if a == "" || b == "" {
		return fmt.Errorf("compare requires two run directories")
	}
	for _, dir := range []string{a, b} {
		if _, err := os.Stat(ScorePath(dir)); err != nil {
			return fmt.Errorf("missing score.json in %s: %w", dir, err)
		}
	}
	return nil
}

func SanitizeVariation(name string) string {
	return strings.TrimSpace(name)
}
