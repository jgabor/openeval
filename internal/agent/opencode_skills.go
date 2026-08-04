package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type openCodeSkill struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

func validateOpenCodeSkills(ctx context.Context, command, workDir string, names, env []string) error {
	if len(names) == 0 {
		return nil
	}

	globalDir, err := os.MkdirTemp("", "openeval-opencode-skills-")
	if err != nil {
		return newOpenCodeRunError("skill", fmt.Sprintf("create isolated skill check: %v", err), "check temporary directory permissions")
	}
	defer func() { _ = os.RemoveAll(globalDir) }()

	global, err := discoverOpenCodeSkills(ctx, command, globalDir, env)
	if err != nil {
		return err
	}
	for _, name := range names {
		if found, ok := findOpenCodeSkill(global, name); ok {
			return newOpenCodeRunError(
				"skill",
				fmt.Sprintf("requested skill %q collides with global discovery at %s", name, found.Location),
				"rename the variation skill or remove the same-named global skill, then retry",
			)
		}
	}

	local, err := discoverOpenCodeSkills(ctx, command, workDir, env)
	if err != nil {
		return err
	}
	for _, name := range names {
		found, ok := findOpenCodeSkill(local, name)
		if !ok {
			return newOpenCodeRunError(
				"skill",
				fmt.Sprintf("requested skill %q is not visible to `opencode debug skill` in %s", name, workDir),
				"check SKILL.md frontmatter and workspace skill permissions, then retry",
			)
		}
		expected := filepath.Join(workDir, ".agents", "skills", name, "SKILL.md")
		if !sameOpenCodeSkillPath(found.Location, expected, workDir) {
			return newOpenCodeRunError(
				"skill",
				fmt.Sprintf("requested skill %q resolved to %s instead of seeded path %s", name, found.Location, expected),
				"remove the colliding OpenCode skill or rename the variation skill, then retry",
			)
		}
	}
	return nil
}

func discoverOpenCodeSkills(ctx context.Context, command, dir string, env []string) ([]openCodeSkill, error) {
	cmd := exec.CommandContext(ctx, command, "debug", "skill")
	cmd.Dir = dir
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		detail := stderr.String()
		if detail == "" {
			detail = err.Error()
		}
		return nil, newOpenCodeRunError(
			"skill",
			fmt.Sprintf("`opencode debug skill` failed in %s: %s", dir, detail),
			"verify OpenCode 1.18.11 is configured and rerun `opencode debug skill` in the workspace",
		)
	}
	var skills []openCodeSkill
	if err := json.Unmarshal(stdout.Bytes(), &skills); err != nil {
		return nil, newOpenCodeRunError(
			"skill",
			fmt.Sprintf("parse `opencode debug skill` output in %s: %v", dir, err),
			"verify OpenCode 1.18.11 returns a JSON skill list",
		)
	}
	return skills, nil
}

func findOpenCodeSkill(skills []openCodeSkill, name string) (openCodeSkill, bool) {
	for _, skill := range skills {
		if skill.Name == name {
			return skill, true
		}
	}
	return openCodeSkill{}, false
}

func sameOpenCodeSkillPath(got, want, base string) bool {
	if !filepath.IsAbs(got) {
		got = filepath.Join(base, got)
	}
	got, _ = filepath.Abs(got)
	want, _ = filepath.Abs(want)
	if resolved, err := filepath.EvalSymlinks(got); err == nil {
		got = resolved
	}
	if resolved, err := filepath.EvalSymlinks(want); err == nil {
		want = resolved
	}
	return filepath.Clean(got) == filepath.Clean(want)
}
