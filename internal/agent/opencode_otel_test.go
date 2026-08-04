package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/runcontext"
	"github.com/jgabor/openeval/internal/scenario"
)

func TestWithOpenCodeNativeOTELPreservesExporterSettingsAndAddsCorrelation(t *testing.T) {
	env, err := withOpenCodeNativeOTEL(
		[]string{
			"OTEL_EXPORTER_OTLP_ENDPOINT=http://inherited:4318",
			"OTEL_EXPORTER_OTLP_HEADERS=Authorization=Bearer%20test",
			"OTEL_RESOURCE_ATTRIBUTES=service.namespace=my%20team,openeval.task_id=stale",
		},
		true,
		"https://collector.example/otel/v1/traces",
		runcontext.Context{
			ScenarioID: "scenario,one",
			Variation:  "with skill",
			TaskID:     "task=one",
			Round:      2,
			TraceID:    "trace/one",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT=https://collector.example/otel",
		"OTEL_EXPORTER_OTLP_HEADERS=Authorization=Bearer%20test",
		"OTEL_RESOURCE_ATTRIBUTES=service.namespace=my%20team,openeval.trace_id=trace%2Fone,openeval.scenario_id=scenario%2Cone,openeval.variation=with%20skill,openeval.task_id=task%3Done,openeval.round=2",
	}
	if !slices.Equal(env, want) {
		t.Fatalf("got %v, want %v", env, want)
	}
}

func TestWithOpenCodeNativeOTELIsExplicitOptIn(t *testing.T) {
	inherited := []string{"OTEL_EXPORTER_OTLP_ENDPOINT=http://inherited:4318"}
	got, err := withOpenCodeNativeOTEL(inherited, false, "not-a-url", runcontext.Context{TraceID: "trace"})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, inherited) {
		t.Fatalf("got %v, want inherited environment unchanged", got)
	}
}

func TestWithOpenCodeNativeOTELReportsInvalidEndpoint(t *testing.T) {
	_, err := withOpenCodeNativeOTEL(nil, true, "not-a-url", runcontext.Context{TraceID: "trace"})
	if err == nil {
		t.Fatal("expected invalid endpoint error")
	}
	for _, want := range []string{"telemetry error", "native OTEL endpoint", "openeval instrument --agent opencode"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q missing %q", err, want)
		}
	}
}

func TestOpenCodeRunPassesOptInNativeOTELToRuntime(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "Authorization=Bearer%20test")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.namespace=existing")
	root := t.TempDir()
	envPath := filepath.Join(root, "otel-env.log")
	stub := filepath.Join(root, "opencode-stub.sh")
	script := fmt.Sprintf(`#!/bin/sh
set -eu
{
  printf 'endpoint=%%s\n' "${OTEL_EXPORTER_OTLP_ENDPOINT-}"
  printf 'headers=%%s\n' "${OTEL_EXPORTER_OTLP_HEADERS-}"
  printf 'resources=%%s\n' "${OTEL_RESOURCE_ATTRIBUTES-}"
} > %q
printf '%%s\n' '{"type":"step_finish","sessionID":"ses-otel","part":{"tokens":{}}}'
`, envPath)
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	workDir := filepath.Join(root, "workspace")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Agents.OpenCode.Command = stub
	cfg.Agents.OpenCode.NativeOTEL = true
	cfg.Telemetry.Endpoint = "https://collector.example/otlp/v1/traces"
	driver, err := New("opencode", cfg)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = driver.Run(context.Background(), Session{
		WorkDir:   workDir,
		Task:      scenario.Task{ID: "edit", Prompt: "edit"},
		Variation: scenario.Variation{},
		Run: runcontext.Context{
			ScenarioID: "scenario",
			Variation:  "baseline",
			TaskID:     "edit",
			Round:      1,
			TraceID:    "trace-id",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	for _, want := range []string{
		"endpoint=https://collector.example/otlp",
		"headers=Authorization=Bearer%20test",
		"resources=service.namespace=existing,openeval.trace_id=trace-id",
		"openeval.scenario_id=scenario",
		"openeval.variation=baseline",
		"openeval.task_id=edit",
		"openeval.round=1",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("native OTEL environment missing %q:\n%s", want, body)
		}
	}
}
