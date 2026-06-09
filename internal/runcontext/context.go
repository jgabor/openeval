package runcontext

import "fmt"

type Context struct {
	ScenarioID string
	Variation  string
	TaskID     string
	Round      int
	TraceID    string
}

func (c Context) Active() bool {
	return c.TraceID != ""
}

func (c Context) Env() []string {
	if !c.Active() {
		return nil
	}
	return []string{
		fmt.Sprintf("OPENEVAL_SCENARIO_ID=%s", c.ScenarioID),
		fmt.Sprintf("OPENEVAL_VARIATION=%s", c.Variation),
		fmt.Sprintf("OPENEVAL_TASK_ID=%s", c.TaskID),
		fmt.Sprintf("OPENEVAL_ROUND=%d", c.Round),
		fmt.Sprintf("OPENEVAL_TRACE_ID=%s", c.TraceID),
	}
}
