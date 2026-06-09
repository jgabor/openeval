package cursor

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRunExportsMaskedSpan(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "openeval")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfg := map[string]any{
		"version": 1,
		"telemetry": map[string]any{
			"protocol": "http/json",
			"endpoint": srv.URL,
			"insecure": true,
		},
		"privacy": map[string]any{
			"mask_prompts": true,
			"mask_secrets": true,
		},
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", dir)

	payload, _ := json.Marshal(map[string]string{
		"hook_event_name": "sessionStart",
		"conversation_id": "conv-1",
		"prompt":          "secret prompt text",
	})
	if err := Run(context.Background(), strings.NewReader(string(payload))); err != nil {
		t.Fatal(err)
	}
	if gotBody == "" {
		t.Fatal("expected OTLP payload")
	}
	if strings.Contains(gotBody, "secret prompt text") {
		t.Fatalf("prompt leaked into export: %s", gotBody)
	}
	if !strings.Contains(gotBody, "[masked]") {
		t.Fatalf("expected masked prompt in export: %s", gotBody)
	}
}
