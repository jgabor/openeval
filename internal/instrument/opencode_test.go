package instrument

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/config"
)

func TestInstallOpenCodeEnablesNativeOTELIdempotently(t *testing.T) {
	cfg := config.Default()
	first, err := Install("opencode", &cfg)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Install("opencode", &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Agents.OpenCode.NativeOTEL {
		t.Fatal("native OpenCode OTEL was not enabled")
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("second install changed output: first=%+v second=%+v", first, second)
	}
	body := strings.Join(first.Messages, "\n")
	for _, want := range []string{
		"http://localhost:4318/v1/traces",
		"OpenEval prompt and secret masking does not apply",
		"export OTEL_EXPORTER_OTLP_ENDPOINT='http://localhost:4318'",
		"OTEL_EXPORTER_OTLP_HEADERS",
		"OTEL_RESOURCE_ATTRIBUTES",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("instrument output missing %q:\n%s", want, body)
		}
	}
}

func TestInstallOpenCodeRejectsInvalidEndpointWithoutEnabling(t *testing.T) {
	cfg := config.Default()
	cfg.Telemetry.Endpoint = "localhost:4318"
	_, err := Install("opencode", &cfg)
	if err == nil {
		t.Fatal("expected invalid endpoint error")
	}
	if cfg.Agents.OpenCode.NativeOTEL {
		t.Fatal("native OTEL enabled despite invalid endpoint")
	}
}
