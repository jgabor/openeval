package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jgabor/openeval/internal/config"
)

type Exporter struct {
	cfg  config.TelemetryConfig
	priv config.PrivacyConfig
}

func New(cfg config.Config) *Exporter {
	return &Exporter{cfg: cfg.Telemetry, priv: cfg.Privacy}
}

func (e *Exporter) EmitSession(ctx context.Context, service, traceID string, costUSD float64, skills []string) error {
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
								"traceId":           traceID,
								"spanId":            spanID(traceID),
								"name":              "openeval.session",
								"startTimeUnixNano": fmt.Sprintf("%d", time.Now().UnixNano()),
								"endTimeUnixNano":   fmt.Sprintf("%d", time.Now().UnixNano()),
								"attributes": []map[string]any{
									{"key": "openeval.cost_usd", "value": map[string]float64{"doubleValue": costUSD}},
									{"key": "openeval.skills", "value": map[string]string{"stringValue": strings.Join(skills, ",")}},
									{"key": "openeval.mask_prompts", "value": map[string]bool{"boolValue": e.priv.MaskPrompts}},
								},
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

func spanID(traceID string) string {
	if len(traceID) >= 16 {
		return traceID[:16]
	}
	return fmt.Sprintf("%016x", traceID)
}
