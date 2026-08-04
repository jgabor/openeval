package agent

import (
	"strings"
	"testing"

	"github.com/jgabor/openeval/internal/config"
)

func TestResolveOpenCodeModelPrecedence(t *testing.T) {
	cfg := config.Default()
	cfg.Agents.OpenCode.Model = "config/model"
	tests := []struct {
		name     string
		user     string
		scenario string
		want     string
	}{
		{name: "user", user: "user/model", scenario: "scenario/model", want: "user/model"},
		{name: "scenario", scenario: "scenario/model", want: "scenario/model"},
		{name: "config", want: "config/model"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveModel("opencode", tt.user, tt.scenario, cfg)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("ResolveModel() = %q, want %q", got, tt.want)
			}
		})
	}

	cfg.Agents.OpenCode.Model = ""
	got, err := ResolveModel("opencode", "", "", cfg)
	if err != nil {
		t.Fatal(err)
	}
	if got != DefaultOpenCodeModel {
		t.Fatalf("ResolveModel() = %q, want default %q", got, DefaultOpenCodeModel)
	}
}

func TestResolveModelRejectsInvalidOpenCodeIdentifier(t *testing.T) {
	for _, model := range []string{"missing-provider", "/model", "provider/", "provider//model", "provider/model/", "provider/model name"} {
		t.Run(model, func(t *testing.T) {
			_, err := ResolveModel("opencode", model, "", config.Default())
			if err == nil || !strings.Contains(err.Error(), "expected provider/model") {
				t.Fatalf("ResolveModel(%q) error = %v, want provider/model syntax error", model, err)
			}
		})
	}
}

func TestResolveModelKeepsCursorAndMockBehavior(t *testing.T) {
	for _, agentName := range []string{"cursor", "mock"} {
		t.Run(agentName+" ignores scenario OpenCode model", func(t *testing.T) {
			got, err := ResolveModel(agentName, "", "scenario/model", config.Default())
			if err != nil {
				t.Fatal(err)
			}
			if got != "" {
				t.Fatalf("ResolveModel() = %q, want no model", got)
			}
		})
		t.Run(agentName+" rejects user model", func(t *testing.T) {
			_, err := ResolveModel(agentName, "user/model", "", config.Default())
			if err == nil || !strings.Contains(err.Error(), "not supported") {
				t.Fatalf("ResolveModel() error = %v, want unsupported model error", err)
			}
		})
	}
}
