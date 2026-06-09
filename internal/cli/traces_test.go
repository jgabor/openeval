package cli

import (
	"bytes"
	"testing"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/score"
)

func TestLookupTraceID(t *testing.T) {
	doc := score.Document{
		ByTask: []score.TaskResult{{
			TaskID: "hello-verify",
			Rounds: []score.RoundResult{{Round: 1, TraceID: "trace-abc"}},
		}},
	}
	got, err := lookupTraceID(doc, "hello-verify", 1)
	if err != nil {
		t.Fatal(err)
	}
	if got != "trace-abc" {
		t.Fatalf("traceID = %q", got)
	}
}

func TestWriteTracesOutput(t *testing.T) {
	var out bytes.Buffer
	cfg := config.Default()
	cfg.Telemetry.Endpoint = "http://localhost:4318/v1/traces"
	writeTracesOutput(&out, "trace-abc", cfg)
	body := out.String()
	for _, want := range []string{"trace_id: trace-abc", "otlp_endpoint: http://localhost:4318/v1/traces", "jaeger_ui:"} {
		if !bytes.Contains([]byte(body), []byte(want)) {
			t.Fatalf("missing %q in %q", want, body)
		}
	}
}
