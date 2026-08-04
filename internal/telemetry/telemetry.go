package telemetry

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jgabor/openeval/internal/config"
	"github.com/jgabor/openeval/internal/runcontext"
)

type Exporter struct {
	cfg  config.TelemetryConfig
	priv config.PrivacyConfig
}

func New(cfg config.Config) *Exporter {
	return &Exporter{cfg: cfg.Telemetry, priv: cfg.Privacy}
}

func (e *Exporter) EmitSession(ctx context.Context, service, traceID string, costUSD float64, skills []string, run runcontext.Context) error {
	attrs := []map[string]any{
		{"key": "openeval.cost_usd", "value": map[string]float64{"doubleValue": costUSD}},
		{"key": "openeval.skills", "value": map[string]string{"stringValue": strings.Join(skills, ",")}},
		{"key": "openeval.mask_prompts", "value": map[string]bool{"boolValue": e.priv.MaskPrompts}},
	}
	if run.Active() {
		attrs = append(attrs,
			map[string]any{"key": "openeval.scenario_id", "value": map[string]string{"stringValue": run.ScenarioID}},
			map[string]any{"key": "openeval.variation", "value": map[string]string{"stringValue": run.Variation}},
			map[string]any{"key": "openeval.task_id", "value": map[string]string{"stringValue": run.TaskID}},
			map[string]any{"key": "openeval.round", "value": map[string]string{"stringValue": fmt.Sprintf("%d", run.Round)}},
		)
	}
	return e.emitSpan(ctx, service, traceID, "openeval.session", attrs)
}

func (e *Exporter) EmitHookEvent(ctx context.Context, traceID, name string, attrs map[string]string) error {
	spanAttrs := []map[string]any{
		{"key": "openeval.mask_prompts", "value": map[string]bool{"boolValue": e.priv.MaskPrompts}},
		{"key": "openeval.mask_secrets", "value": map[string]bool{"boolValue": e.priv.MaskSecrets}},
	}
	for k, v := range attrs {
		spanAttrs = append(spanAttrs, map[string]any{
			"key":   k,
			"value": map[string]string{"stringValue": v},
		})
	}
	return e.emitSpan(ctx, "openeval-hook-cursor", traceID, name, spanAttrs)
}

func (e *Exporter) emitSpan(ctx context.Context, service, traceID, name string, attrs []map[string]any) error {
	if e.cfg.Endpoint == "" {
		return nil
	}
	payload := map[string]any{
		"resourceSpans": []map[string]any{
			{
				"resource": map[string]any{
					"attributes": []map[string]any{
						{"key": "service.name", "value": map[string]string{"stringValue": service}},
					},
				},
				"scopeSpans": []map[string]any{
					{
						"spans": []map[string]any{
							{
								"traceId":           NormalizeTraceID(traceID),
								"spanId":            spanID(traceID),
								"name":              name,
								"startTimeUnixNano": fmt.Sprintf("%d", time.Now().UnixNano()),
								"endTimeUnixNano":   fmt.Sprintf("%d", time.Now().UnixNano()),
								"attributes":        attrs,
							},
						},
					},
				},
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("otlp export status %s", resp.Status)
	}
	return nil
}

var hexTrace = regexp.MustCompile(`^[0-9a-fA-F]{32}$`)

func NormalizeTraceID(id string) string {
	id = strings.TrimSpace(id)
	if hexTrace.MatchString(id) {
		return strings.ToLower(id)
	}
	sum := sha256.Sum256([]byte(id))
	return hex.EncodeToString(sum[:16])
}

func RandomTraceID() string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(sum[:16])
}

func OpenCodeOTLPBaseEndpoint(traceEndpoint string) (string, error) {
	raw := strings.TrimSpace(traceEndpoint)
	if raw == "" {
		return "", fmt.Errorf("telemetry endpoint is empty")
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("telemetry endpoint %q must be an absolute HTTP(S) URL", traceEndpoint)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("telemetry endpoint %q must use http or https", traceEndpoint)
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("telemetry endpoint %q must not contain a query or fragment", traceEndpoint)
	}
	path := strings.TrimRight(parsed.Path, "/")
	path = strings.TrimSuffix(path, "/v1/traces")
	parsed.Path = path
	parsed.RawPath = ""
	return strings.TrimRight(parsed.String(), "/"), nil
}

func OpenCodeOTLPTraceEndpoint(baseEndpoint string) string {
	return strings.TrimRight(baseEndpoint, "/") + "/v1/traces"
}

func MaskSecrets(value string) string {
	if strings.Contains(value, "sk-") || strings.Contains(value, "Bearer ") || strings.Contains(value, "api_key=") {
		return "[masked]"
	}
	return value
}

func spanID(traceID string) string {
	normalized := NormalizeTraceID(traceID)
	if len(normalized) >= 16 {
		return normalized[:16]
	}
	return fmt.Sprintf("%016x", normalized)
}
