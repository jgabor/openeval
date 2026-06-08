package report

import (
	"fmt"
	"strings"

	"github.com/jgabor/openeval/internal/paths"
	"github.com/jgabor/openeval/internal/score"
)

func LoadRun(runDir string) (score.Document, error) {
	return score.Load(paths.ScorePath(runDir))
}

func Format(doc score.Document) string {
	var b strings.Builder
	fmt.Fprintf(&b, "scenario: %s\n", doc.ScenarioID)
	fmt.Fprintf(&b, "agent: %s\n", doc.Agent)
	if doc.Variation != "" && doc.Variation != "default" {
		fmt.Fprintf(&b, "variation: %s\n", doc.Variation)
	}
	fmt.Fprintf(&b, "rounds: %d\n\n", doc.Rounds)
	fmt.Fprintf(&b, "pass@1: %.2f  (%d/%d tasks passed on first attempt)\n", doc.Summary.PassAt1, int(doc.Summary.PassAt1*float64(doc.Summary.TasksTotal)), doc.Summary.TasksTotal)
	fmt.Fprintf(&b, "pass@%d: %.2f  (%d/%d tasks passed within %d attempts)\n", doc.Rounds, doc.Summary.PassAt3, doc.Summary.TasksPassed, doc.Summary.TasksTotal, doc.Rounds)
	fmt.Fprintf(&b, "cost:   $%.2f total  ($%.2f per passed task)\n\n", doc.Summary.CostUSDTotal, doc.Summary.CostUSDPerPassed)
	fmt.Fprintf(&b, "tasks:\n")
	for _, t := range doc.ByTask {
		passK := t.PassAtK[fmt.Sprintf("%d", doc.Rounds)]
		var taskCost float64
		for _, r := range t.Rounds {
			taskCost += r.CostUSD
		}
		status := strings.ToUpper(t.Verifier)
		fmt.Fprintf(&b, "  %-18s %-4s pass@%d=%.2f  cost=$%.2f\n", t.TaskID, status, doc.Rounds, passK, taskCost)
	}
	fmt.Fprintf(&b, "\ntraces: %d sessions, OTLP service openeval-agent\n", doc.Telemetry.Sessions)
	return b.String()
}
