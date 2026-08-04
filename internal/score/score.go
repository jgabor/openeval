package score

import (
	"encoding/json"
	"fmt"
	"os"
)

const Schema = "openeval.score.v1"

type Document struct {
	Schema     string           `json:"schema"`
	ScenarioID string           `json:"scenario_id"`
	Agent      string           `json:"agent"`
	Model      string           `json:"model,omitempty"`
	Variation  string           `json:"variation,omitempty"`
	Rounds     int              `json:"rounds"`
	Tasks      int              `json:"tasks"`
	Summary    Summary          `json:"summary"`
	ByTask     []TaskResult     `json:"by_task"`
	Telemetry  TelemetrySummary `json:"telemetry"`
}

type Summary struct {
	PassAt1           float64 `json:"pass_at_1"`
	PassAt3           float64 `json:"pass_at_3"`
	TasksPassed       int     `json:"tasks_passed"`
	TasksTotal        int     `json:"tasks_total"`
	CostUSDTotal      float64 `json:"cost_usd_total"`
	CostUSDPerPassed  float64 `json:"cost_usd_per_passed_task"`
	TokensInputTotal  int     `json:"tokens_input_total"`
	TokensOutputTotal int     `json:"tokens_output_total"`
}

type TaskResult struct {
	TaskID   string             `json:"task_id"`
	Verifier string             `json:"verifier"`
	Rounds   []RoundResult      `json:"rounds"`
	PassAtK  map[string]float64 `json:"pass_at_k"`
}

type RoundResult struct {
	Round    int     `json:"round"`
	Verifier string  `json:"verifier"`
	CostUSD  float64 `json:"cost_usd"`
	TraceID  string  `json:"trace_id"`
}

type TelemetrySummary struct {
	SkillsActive   []string           `json:"skills_active"`
	CostBySkillUSD map[string]float64 `json:"cost_by_skill_usd"`
	Sessions       int                `json:"sessions"`
}

func Load(path string) (Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Document{}, err
	}
	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return Document{}, fmt.Errorf("parse score %s: %w", path, err)
	}
	return doc, nil
}

func Save(path string, doc Document) error {
	doc.Schema = Schema
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func ComputePassAtK(rounds []RoundResult, k int) float64 {
	if len(rounds) == 0 {
		return 0
	}
	limit := k
	if limit > len(rounds) {
		limit = len(rounds)
	}
	for i := 0; i < limit; i++ {
		if rounds[i].Verifier == "pass" {
			return 1
		}
	}
	return 0
}

func BuildSummary(byTask []TaskResult, rounds int) Summary {
	total := len(byTask)
	passed := 0
	var costTotal float64
	var tokensIn, tokensOut int
	pass1 := 0.0
	passK := 0.0
	k := rounds
	if k < 1 {
		k = 3
	}
	for _, t := range byTask {
		if t.PassAtK[fmt.Sprintf("%d", k)] >= 1 {
			passed++
		}
		pass1 += t.PassAtK["1"]
		passK += t.PassAtK[fmt.Sprintf("%d", k)]
		for _, r := range t.Rounds {
			costTotal += r.CostUSD
		}
	}
	if total > 0 {
		pass1 /= float64(total)
		passK /= float64(total)
	}
	perPassed := costTotal
	if passed > 0 {
		perPassed = costTotal / float64(passed)
	}
	return Summary{
		PassAt1:           pass1,
		PassAt3:           passK,
		TasksPassed:       passed,
		TasksTotal:        total,
		CostUSDTotal:      costTotal,
		CostUSDPerPassed:  perPassed,
		TokensInputTotal:  tokensIn,
		TokensOutputTotal: tokensOut,
	}
}
