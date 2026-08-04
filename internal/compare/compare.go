package compare

import (
	"fmt"
	"strings"

	"github.com/jgabor/openeval/internal/paths"
	"github.com/jgabor/openeval/internal/score"
)

func Run(dirA, dirB string) (string, error) {
	if err := paths.ValidateCompareDirs(dirA, dirB); err != nil {
		return "", err
	}
	a, err := score.Load(paths.ScorePath(dirA))
	if err != nil {
		return "", err
	}
	b, err := score.Load(paths.ScorePath(dirB))
	if err != nil {
		return "", err
	}
	if a.ScenarioID != b.ScenarioID {
		return "", fmt.Errorf("scenario mismatch: %s vs %s", a.ScenarioID, b.ScenarioID)
	}
	warnAgent := a.Agent != b.Agent
	labelA := a.Variation
	if labelA == "" {
		labelA = "a"
	}
	labelB := b.Variation
	if labelB == "" {
		labelB = "b"
	}
	var out strings.Builder
	fmt.Fprintf(&out, "comparing: %s to %s\n", labelA, labelB)
	fmt.Fprintf(&out, "scenario:  %s\n", a.ScenarioID)
	fmt.Fprintf(&out, "agent:     %s\n", a.Agent)
	if warnAgent {
		fmt.Fprintf(&out, "warning:   agent mismatch (%s vs %s)\n", a.Agent, b.Agent)
	}
	if a.Model != "" || b.Model != "" {
		fmt.Fprintf(&out, "model:     %s vs %s\n", knownModel(a.Model), knownModel(b.Model))
	}
	if a.Model != "" && b.Model != "" && a.Model != b.Model {
		fmt.Fprintf(&out, "warning:   model mismatch (%s vs %s)\n", a.Model, b.Model)
	}
	fmt.Fprintln(&out)
	fmt.Fprintf(&out, "%24s %24s %8s\n", labelA, labelB, "delta")
	fmt.Fprintf(&out, "%-24s %24.2f %24.2f %8.2f\n", "pass@1", a.Summary.PassAt1, b.Summary.PassAt1, b.Summary.PassAt1-a.Summary.PassAt1)
	fmt.Fprintf(&out, "%-24s %24.2f %24.2f %8.2f\n", fmt.Sprintf("pass@%d", a.Rounds), a.Summary.PassAt3, b.Summary.PassAt3, b.Summary.PassAt3-a.Summary.PassAt3)
	fmt.Fprintf(&out, "%-24s %24.2f %24.2f %+8.2f\n", "cost_usd_total", a.Summary.CostUSDTotal, b.Summary.CostUSDTotal, b.Summary.CostUSDTotal-a.Summary.CostUSDTotal)
	fmt.Fprintf(&out, "%-24s %24.2f %24.2f %+8.2f\n", "cost_per_passed", a.Summary.CostUSDPerPassed, b.Summary.CostUSDPerPassed, b.Summary.CostUSDPerPassed-a.Summary.CostUSDPerPassed)
	fmt.Fprintf(&out, "\nresults:\n  logs:\n    %s\n    %s\n", paths.ScorePath(dirA), paths.ScorePath(dirB))
	return out.String(), nil
}

func knownModel(model string) string {
	if model == "" {
		return "unknown"
	}
	return model
}
