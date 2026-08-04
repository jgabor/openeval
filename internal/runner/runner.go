package runner

import (
	"context"
	"fmt"

	"github.com/jgabor/openeval/internal/agent"
	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/paths"
	"github.com/jgabor/openeval/internal/runcontext"
	"github.com/jgabor/openeval/internal/scenario"
	"github.com/jgabor/openeval/internal/score"
	"github.com/jgabor/openeval/internal/telemetry"
	"github.com/jgabor/openeval/internal/verifier"
	"github.com/jgabor/openeval/internal/workspace"
)

type Options struct {
	Scenario  string
	Agent     string
	Model     string
	Variation string
	Rounds    int
	Out       string
	Image     string
}

type Result struct {
	RunDir string
	Score  score.Document
}

func Run(ctx context.Context, opts Options) (Result, error) {
	if opts.Rounds <= 0 {
		opts.Rounds = 3
	}
	cfg, err := config.Load()
	if err != nil {
		return Result{}, err
	}
	sc, err := scenario.Load(opts.Scenario, cfg)
	if err != nil {
		return Result{}, err
	}
	variationName := paths.SanitizeVariation(opts.Variation)
	if variationName == "" {
		variationName = "default"
	}
	variation, err := sc.Variation(variationName)
	if err != nil {
		return Result{}, err
	}
	model, err := agent.ResolveModel(opts.Agent, opts.Model, sc.Model, cfg)
	if err != nil {
		return Result{}, err
	}
	runDir, err := paths.ResolveRunDir(sc.ID, variationName, opts.Out)
	if err != nil {
		return Result{}, err
	}
	driver, err := agent.New(opts.Agent, cfg)
	if err != nil {
		return Result{}, err
	}
	exporter := telemetry.New(cfg)
	byTask := make([]score.TaskResult, 0, len(sc.Tasks))
	sessions := 0
	costBySkill := map[string]float64{}
	for _, skill := range variation.Skills {
		costBySkill[skill] = 0
	}
	for _, task := range sc.Tasks {
		rounds := make([]score.RoundResult, 0, opts.Rounds)
		for round := 1; round <= opts.Rounds; round++ {
			workDir := workspace.RoundDir(runDir, task.ID, round)
			if err := workspace.Seed(sc, workDir); err != nil {
				return Result{}, fmt.Errorf("task %s round %d workspace: %w", task.ID, round, err)
			}
			if err := workspace.SeedSkills(workDir, variation.Skills, cfg); err != nil {
				return Result{}, fmt.Errorf("task %s round %d skills: %w", task.ID, round, err)
			}
			pluginDirs, err := workspace.SkillPluginDirs(variation.Skills, cfg)
			if err != nil {
				return Result{}, fmt.Errorf("task %s round %d plugin dirs: %w", task.ID, round, err)
			}
			traceID := telemetry.RandomTraceID()
			sess := agent.Session{
				WorkDir:    workDir,
				Agent:      opts.Agent,
				Model:      model,
				Variation:  variation,
				Task:       task,
				Round:      round,
				PluginDirs: pluginDirs,
				Run: runcontext.Context{
					ScenarioID: sc.ID,
					Variation:  variationName,
					TaskID:     task.ID,
					Round:      round,
					TraceID:    traceID,
				},
			}
			cost, traceID, err := driver.Run(ctx, sess)
			if err != nil {
				return Result{}, fmt.Errorf("task %s round %d: %w", task.ID, round, err)
			}
			verdict, err := verifier.Run(ctx, sc, task, workDir, variation)
			if err != nil {
				return Result{}, err
			}
			_ = exporter.EmitSession(ctx, "openeval-agent", traceID, cost, variation.Skills, sess.Run)
			rounds = append(rounds, score.RoundResult{
				Round:    round,
				Verifier: verdict,
				CostUSD:  cost,
				TraceID:  traceID,
			})
			sessions++
			for _, skill := range variation.Skills {
				costBySkill[skill] += cost
			}
		}
		passAtK := map[string]float64{
			"1": score.ComputePassAtK(rounds, 1),
		}
		passAtK[fmt.Sprintf("%d", opts.Rounds)] = score.ComputePassAtK(rounds, opts.Rounds)
		final := "fail"
		if passAtK[fmt.Sprintf("%d", opts.Rounds)] >= 1 {
			final = "pass"
		}
		byTask = append(byTask, score.TaskResult{
			TaskID:   task.ID,
			Verifier: final,
			Rounds:   rounds,
			PassAtK:  passAtK,
		})
	}
	doc := score.Document{
		ScenarioID: sc.ID,
		Agent:      opts.Agent,
		Variation:  omitDefaultVariation(variationName),
		Rounds:     opts.Rounds,
		Tasks:      len(sc.Tasks),
		ByTask:     byTask,
		Telemetry: score.TelemetrySummary{
			SkillsActive:   variation.Skills,
			CostBySkillUSD: costBySkill,
			Sessions:       sessions,
		},
	}
	doc.Summary = score.BuildSummary(byTask, opts.Rounds)
	if err := score.Save(paths.ScorePath(runDir), doc); err != nil {
		return Result{}, err
	}
	return Result{RunDir: runDir, Score: doc}, nil
}

func omitDefaultVariation(name string) string {
	if name == "default" {
		return ""
	}
	return name
}
