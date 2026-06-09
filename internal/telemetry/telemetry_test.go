package telemetry

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/runcontext"
)

func TestEmitSessionIncludesRunContextAndNormalizedTraceID(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := config.Default()
	cfg.Telemetry.Endpoint = srv.URL
	exporter := New(cfg)
	traceID := "corr-trace-abc"
	run := runcontext.Context{
		ScenarioID: "example-fixtures",
		Variation:  "default",
		TaskID:     "edit-file",
		Round:      1,
		TraceID:    traceID,
	}
	if err := exporter.EmitSession(context.Background(), "openeval-agent", traceID, 0.5, nil, run); err != nil {
		t.Fatal(err)
	}
	normalized := NormalizeTraceID(traceID)
	if !strings.Contains(gotBody, normalized) {
		t.Fatalf("expected normalized traceId %q in %s", normalized, gotBody)
	}
	for _, want := range []string{"openeval.scenario_id", "example-fixtures", "openeval.task_id", "edit-file"} {
		if !strings.Contains(gotBody, want) {
			t.Fatalf("missing %q in %s", want, gotBody)
		}
	}
}
