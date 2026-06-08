package verifier

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jgabor/openeval/internal/scenario"
)

func Run(ctx context.Context, sc scenario.Scenario, task scenario.Task, workDir string, variation scenario.Variation) (string, error) {
	if task.Verifier.Type != "script" {
		return "fail", fmt.Errorf("unsupported verifier type %q", task.Verifier.Type)
	}
	script := task.Verifier.Run
	if !filepath.IsAbs(script) {
		script = filepath.Join(sc.SourceDir(), script)
	}
	cmd := exec.CommandContext(ctx, script)
	cmd.Dir = workDir
	cmd.Env = os.Environ()
	for k, v := range variation.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	if err := cmd.Run(); err != nil {
		return "fail", nil
	}
	return "pass", nil
}
