package cursor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/telemetry"
)

func Run(ctx context.Context, r io.Reader) error {
	raw, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return nil
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("parse cursor hook input: %w", err)
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	event := stringField(payload, "hook_event_name")
	if event == "" {
		event = "cursor.hook"
	}
	traceID := strings.TrimSpace(os.Getenv("OPENEVAL_TRACE_ID"))
	if traceID == "" {
		traceID = firstString(payload, "conversation_id", "session_id", "generation_id")
	}
	if traceID == "" {
		traceID = telemetry.RandomTraceID()
	}
	attrs := map[string]string{
		"cursor.hook_event": event,
	}
	for _, key := range []string{"conversation_id", "generation_id", "tool_name", "workspace_roots"} {
		if v := stringField(payload, key); v != "" {
			attrs[key] = redact(cfg.Privacy, key, v)
		}
	}
	for _, key := range []string{"OPENEVAL_SCENARIO_ID", "OPENEVAL_VARIATION", "OPENEVAL_TASK_ID", "OPENEVAL_ROUND"} {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			attrs[strings.ToLower(key)] = v
		}
	}
	if prompt := stringField(payload, "prompt"); prompt != "" {
		attrs["prompt"] = redact(cfg.Privacy, "prompt", prompt)
	}
	exporter := telemetry.New(cfg)
	return exporter.EmitHookEvent(ctx, traceID, event, attrs)
}

func stringField(payload map[string]any, key string) string {
	v, ok := payload[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

func firstString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if v := stringField(payload, key); v != "" {
			return v
		}
	}
	return ""
}

func redact(priv config.PrivacyConfig, key, value string) string {
	lower := strings.ToLower(key)
	if priv.MaskPrompts && (lower == "prompt" || strings.Contains(lower, "prompt")) {
		return "[masked]"
	}
	if priv.MaskSecrets {
		return telemetry.MaskSecrets(value)
	}
	return value
}
