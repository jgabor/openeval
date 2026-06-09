//go:build mage

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

const (
	golangciLintInstall = "go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.1"
	govulncheckInstall  = "go install golang.org/x/vuln/cmd/govulncheck@v1.3.0"
	mainPackage         = "./cmd/openeval"
	binaryPath          = "bin/openeval"
)

var Default = Install

// Build compiles the openeval binary into bin/.
func Build() error {
	if err := os.MkdirAll("bin", 0o755); err != nil {
		return fmt.Errorf("create bin: %w", err)
	}
	return run("go", "build", "-o", binaryPath, mainPackage)
}

// Install builds and installs openeval to GOBIN.
func Install() error {
	return run("go", "install", mainPackage)
}

// Test runs the Go test suite.
func Test() error {
	return run("go", "test", "./...")
}

// Tidy normalizes module dependencies.
func Tidy() error {
	return run("go", "mod", "tidy")
}

// TidyCheck verifies go.mod and go.sum are already tidy.
func TidyCheck() error {
	return run("go", "mod", "tidy", "-diff")
}

// Vet runs go vet.
func Vet() error {
	return run("go", "vet", "./...")
}

// Lint runs golangci-lint.
func Lint() error {
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		return fmt.Errorf("golangci-lint not found; install it with `%s`", golangciLintInstall)
	}
	return run("golangci-lint", "run", "./...")
}

// Vuln runs govulncheck.
func Vuln() error {
	if _, err := exec.LookPath("govulncheck"); err != nil {
		return fmt.Errorf("govulncheck not found; install it with `%s`", govulncheckInstall)
	}
	return run("govulncheck", "./...")
}

// Check runs verification gates used in CI.
func Check() error {
	steps := []struct {
		name string
		run  func() error
	}{
		{name: "tidy", run: TidyCheck},
		{name: "test", run: Test},
		{name: "vet", run: Vet},
		{name: "lint", run: Lint},
		{name: "vuln", run: Vuln},
	}
	for _, step := range steps {
		if err := step.run(); err != nil {
			return fmt.Errorf("%s failed: %w", step.name, err)
		}
	}
	return nil
}

func run(name string, args ...string) error {
	command := exec.Command(name, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	if err := command.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("%s exited with status %d", name, exitErr.ExitCode())
		}
		return err
	}
	return nil
}
