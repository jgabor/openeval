package demo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jgabor/openeval/internal/agent"
	"github.com/jgabor/openeval/internal/compare"
	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/doctor"
	"github.com/jgabor/openeval/internal/runner"
	"github.com/jgabor/openeval/internal/scenario"
)

type Options struct {
	Scenario string
	Agent    string
	Rounds   int
	Out      string
	DryRun   bool
	Model    string
}

type Plan struct {
	Scenario    string
	Agent       string
	Rounds      int
	Model       string
	Root        string
	BaselineDir string
	SkillDir    string
}

type Result struct {
	Plan       Plan
	Baseline   runner.Result
	Skill      runner.Result
	Comparison string
	DryRun     bool
}

func Run(ctx context.Context, opts Options) (Result, error) {
	opts = normalizeOptions(opts)
	model, err := resolveModel(opts)
	if err != nil {
		return Result{}, err
	}
	opts.Model = model
	plan, err := buildPlan(opts)
	if err != nil {
		return Result{}, err
	}
	if opts.DryRun {
		return Result{Plan: plan, DryRun: true}, nil
	}

	if opts.Agent != "mock" {
		diagnosis := doctor.Run(ctx, opts.Agent, plan.Model)
		if diagnosis.ExitCode != 0 {
			return Result{}, diagnosisError(diagnosis)
		}
	}
	if err := createPlanRoot(plan); err != nil {
		return Result{}, err
	}

	baseline, err := runner.Run(ctx, runner.Options{
		Scenario:  plan.Scenario,
		Agent:     plan.Agent,
		Variation: "baseline",
		Rounds:    plan.Rounds,
		Out:       plan.BaselineDir,
		Model:     plan.Model,
	})
	if err != nil {
		return Result{}, fmt.Errorf("baseline run failed; retained evidence root %s: %w", plan.Root, err)
	}
	skill, err := runner.Run(ctx, runner.Options{
		Scenario:  plan.Scenario,
		Agent:     plan.Agent,
		Variation: "with-demo-skill",
		Rounds:    plan.Rounds,
		Out:       plan.SkillDir,
		Model:     plan.Model,
	})
	if err != nil {
		return Result{}, fmt.Errorf("skill run failed; retained baseline %s: %w", baseline.RunDir, err)
	}
	comparison, err := compare.Run(baseline.RunDir, skill.RunDir)
	if err != nil {
		return Result{}, fmt.Errorf("compare retained runs %s and %s: %w", baseline.RunDir, skill.RunDir, err)
	}
	return Result{Plan: plan, Baseline: baseline, Skill: skill, Comparison: comparison}, nil
}

func (p Plan) Format() string {
	commands := [][]string{
		{"openeval", "doctor", "--agent", p.Agent},
		{"openeval", "run", "--scenario", p.Scenario, "--variation", "baseline", "--agent", p.Agent, "--rounds", fmt.Sprintf("%d", p.Rounds), "--out", p.BaselineDir},
		{"openeval", "run", "--scenario", p.Scenario, "--variation", "with-demo-skill", "--agent", p.Agent, "--rounds", fmt.Sprintf("%d", p.Rounds), "--out", p.SkillDir},
		{"openeval", "compare", p.BaselineDir, p.SkillDir},
	}
	if p.Model != "" {
		commands[0] = append(commands[0], "--model", p.Model)
		commands[1] = append(commands[1], "--model", p.Model)
		commands[2] = append(commands[2], "--model", p.Model)
	}
	if p.Agent == "mock" {
		commands = commands[1:]
	}
	var out strings.Builder
	fmt.Fprintf(&out, "demo evidence root: %s\n", p.Root)
	for i, command := range commands {
		fmt.Fprintf(&out, "%d. %s\n", i+1, formatCommand(command))
	}
	return out.String()
}

func resolveModel(opts Options) (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}
	sc, err := scenario.Load(opts.Scenario, cfg)
	if err != nil {
		return "", err
	}
	return agent.ResolveModel(opts.Agent, opts.Model, sc.Model, cfg)
}

func normalizeOptions(opts Options) Options {
	if opts.Scenario == "" {
		opts.Scenario = "example-fixtures"
	}
	if opts.Agent == "" {
		opts.Agent = "opencode"
	}
	if opts.Rounds <= 0 {
		opts.Rounds = 1
	}
	return opts
}

func buildPlan(opts Options) (Plan, error) {
	base := opts.Out
	if base == "" {
		base = filepath.Join("scenarios", scenarioSlug(opts.Scenario), "demos")
	}
	absBase, err := filepath.Abs(base)
	if err != nil {
		return Plan{}, err
	}
	id, err := planID()
	if err != nil {
		return Plan{}, err
	}
	root := filepath.Join(absBase, id)
	return Plan{
		Scenario:    opts.Scenario,
		Agent:       opts.Agent,
		Rounds:      opts.Rounds,
		Model:       opts.Model,
		Root:        root,
		BaselineDir: filepath.Join(root, "baseline"),
		SkillDir:    filepath.Join(root, "with-demo-skill"),
	}, nil
}

func createPlanRoot(plan Plan) error {
	if err := os.MkdirAll(filepath.Dir(plan.Root), 0o755); err != nil {
		return fmt.Errorf("create demo output parent: %w", err)
	}
	if err := os.Mkdir(plan.Root, 0o755); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("refusing to replace existing demo evidence %s; rerun to allocate a new path", plan.Root)
		}
		return fmt.Errorf("create demo evidence root: %w", err)
	}
	return nil
}

func planID() (string, error) {
	random := make([]byte, 4)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate demo id: %w", err)
	}
	return time.Now().UTC().Format("20060102T150405.000000000Z") + "-" + hex.EncodeToString(random), nil
}

func scenarioSlug(value string) string {
	value = strings.TrimSuffix(filepath.Base(value), filepath.Ext(value))
	var slug strings.Builder
	for _, r := range strings.ToLower(value) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			slug.WriteRune(r)
		} else {
			slug.WriteByte('-')
		}
	}
	value = strings.Trim(slug.String(), "-")
	if value == "" {
		return "scenario"
	}
	return value
}

func diagnosisError(report doctor.Report) error {
	var details []string
	for _, check := range report.Checks {
		if check.Status != doctor.StatusFail {
			continue
		}
		detail := check.ID + ": " + check.Summary
		if check.Remediation != "" {
			detail += "; fix: " + check.Remediation
		}
		details = append(details, detail)
	}
	return fmt.Errorf("doctor failed before demo execution: %s", strings.Join(details, " | "))
}

func formatCommand(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = shellQuote(arg)
	}
	return strings.Join(quoted, " ")
}

func shellQuote(value string) string {
	if value != "" && strings.IndexFunc(value, func(r rune) bool {
		return !shellSafeRune(r)
	}) == -1 {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func shellSafeRune(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("-._/:=", r)
}
