package scenario_test

import (
	"os"
	"strings"
	"testing"
)

const (
	composePath = "examples/otel/compose.yaml"
	smokePath   = "examples/otel/smoke.sh"
	jaegerImage = "cr.jaegertracing.io/jaegertracing/jaeger:2.20.0@sha256:46a886260e04002d8f45e213fc39063fa11a50446048fdaa64786fc0840cb9f8"
)

func TestOTelComposeStaticContract(t *testing.T) {
	chdirRepoRoot(t)
	assertFileContains(t, composePath, []string{
		"image: " + jaegerImage,
		`"127.0.0.1:4318:4318"`,
		`"127.0.0.1:16686:16686"`,
		"healthcheck:",
		"http://localhost:13133/status",
		"timeout: 2s",
		"retries: 30",
	})
}

func TestOTelSmokeAndREADMEStaticContract(t *testing.T) {
	chdirRepoRoot(t)
	assertFileContains(t, smokePath, []string{
		"docker compose -f \"$COMPOSE_FILE\" up -d --wait --wait-timeout 60",
		"http://localhost:4318/v1/traces",
		"score.json",
		"hashlib.sha256(source.encode()).hexdigest()[:32]",
		"http://localhost:16686/api/traces/",
		"deadline = time.monotonic() + 30",
		"openeval.session",
		"openeval-agent",
		"http://localhost:16686/trace/",
		"docker compose -f \"$COMPOSE_FILE\" down",
	})
	assertFileContains(t, "README.md", []string{
		composePath,
		smokePath,
		jaegerImage,
		"docker compose -f examples/otel/compose.yaml up -d --wait --wait-timeout 60",
		"http://localhost:4318/v1/traces",
		"http://localhost:16686",
		"docker compose -f examples/otel/compose.yaml down",
		"amd64, arm64, s390x, and ppc64le",
		"port 4318 or 16686 is allocated",
		"keeps traces only in memory",
	})
}

func assertFileContains(t *testing.T, path string, wants []string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range wants {
		if !strings.Contains(text, want) {
			t.Errorf("%s does not contain %q", path, want)
		}
	}
}
