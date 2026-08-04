package runcontext

import (
	"fmt"
	"sort"
	"strings"
)

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

// MergeEnvironment returns one deterministic environment with later layers
// taking precedence over earlier layers.
func MergeEnvironment(layers ...[]string) []string {
	values := make(map[string]string)
	for _, layer := range layers {
		for _, entry := range layer {
			key, _, ok := strings.Cut(entry, "=")
			if !ok || key == "" {
				continue
			}
			values[key] = entry
		}
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	env := make([]string, 0, len(keys))
	for _, key := range keys {
		env = append(env, values[key])
	}
	return env
}

func FromMap(values map[string]string) []string {
	env := make([]string, 0, len(values))
	for key, value := range values {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}

// Environment applies the runtime precedence contract: inherited values are
// overridden by variation values, and reserved OpenEval run values are final.
func Environment(inherited []string, variation map[string]string, run Context) []string {
	return MergeEnvironment(inherited, FromMap(variation), run.Env())
}

func LookupEnvironment(env []string, key string) (string, bool) {
	for _, entry := range env {
		entryKey, value, ok := strings.Cut(entry, "=")
		if ok && entryKey == key {
			return value, true
		}
	}
	return "", false
}
