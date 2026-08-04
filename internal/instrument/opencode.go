package instrument

import (
	"fmt"
	"strings"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/telemetry"
)

func installOpenCode(cfg *config.Config) (Result, error) {
	baseEndpoint, err := telemetry.OpenCodeOTLPBaseEndpoint(cfg.Telemetry.Endpoint)
	if err != nil {
		return Result{}, fmt.Errorf("configure OpenCode native OTEL: %w", err)
	}
	cfg.Agents.OpenCode.NativeOTEL = true
	return Result{
		Agent: "opencode",
		Messages: []string{
			"native OpenCode OTEL enabled as an explicit opt-in",
			"native trace receiver: " + telemetry.OpenCodeOTLPTraceEndpoint(baseEndpoint),
			"privacy: OpenCode generates native payloads; OpenEval prompt and secret masking does not apply to them",
			"external OpenCode sessions: export OTEL_EXPORTER_OTLP_ENDPOINT=" + shellQuote(baseEndpoint),
			"existing OTEL_EXPORTER_OTLP_HEADERS and OTEL_RESOURCE_ATTRIBUTES remain in effect; harness runs append OpenEval correlation attributes",
		},
	}, nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
