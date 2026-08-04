package agent

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/jgabor/openeval/internal/runcontext"
	"github.com/jgabor/openeval/internal/telemetry"
)

var openCodeCorrelationAttributes = map[string]struct{}{
	"openeval.trace_id":    {},
	"openeval.scenario_id": {},
	"openeval.variation":   {},
	"openeval.task_id":     {},
	"openeval.round":       {},
}

func withOpenCodeNativeOTEL(env []string, enabled bool, traceTarget string, run runcontext.Context) ([]string, error) {
	if !enabled {
		return env, nil
	}
	baseEndpoint, err := telemetry.OpenCodeOTLPBaseEndpoint(traceTarget)
	if err != nil {
		return nil, newOpenCodeRunError(
			"telemetry",
			fmt.Sprintf("native OTEL endpoint: %v", err),
			"set telemetry.endpoint to an OTLP HTTP trace endpoint or rerun `openeval instrument --agent opencode`",
		)
	}
	native := map[string]string{"OTEL_EXPORTER_OTLP_ENDPOINT": baseEndpoint}
	if run.Active() {
		existing, _ := runcontext.LookupEnvironment(env, "OTEL_RESOURCE_ATTRIBUTES")
		native["OTEL_RESOURCE_ATTRIBUTES"] = openCodeResourceAttributes(existing, run)
	}
	return runcontext.MergeEnvironment(env, runcontext.FromMap(native)), nil
}

func openCodeResourceAttributes(existing string, run runcontext.Context) string {
	attributes := make([]string, 0, 10)
	for _, attribute := range strings.Split(existing, ",") {
		attribute = strings.TrimSpace(attribute)
		if attribute == "" {
			continue
		}
		key, _, ok := strings.Cut(attribute, "=")
		decoded, err := url.QueryUnescape(key)
		if ok && err == nil {
			if _, reserved := openCodeCorrelationAttributes[decoded]; reserved {
				continue
			}
		}
		attributes = append(attributes, attribute)
	}
	for _, attribute := range []struct {
		key   string
		value string
	}{
		{key: "openeval.trace_id", value: run.TraceID},
		{key: "openeval.scenario_id", value: run.ScenarioID},
		{key: "openeval.variation", value: run.Variation},
		{key: "openeval.task_id", value: run.TaskID},
		{key: "openeval.round", value: strconv.Itoa(run.Round)},
	} {
		attributes = append(attributes, attribute.key+"="+encodeOTELAttribute(attribute.value))
	}
	return strings.Join(attributes, ",")
}

func encodeOTELAttribute(value string) string {
	return strings.ReplaceAll(url.QueryEscape(value), "+", "%20")
}
