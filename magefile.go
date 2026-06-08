//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var Default = Install

func Build() error {
	return sh.Run("go", "build", "-o", "bin/openeval", "./cmd/openeval")
}

func Install() error {
	mg.Deps(Build)
	bin, err := installDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(bin, 0o755); err != nil {
		return err
	}
	dst := filepath.Join(bin, "openeval")
	return sh.Copy(dst, "bin/openeval")
}

func installDir() (string, error) {
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		return filepath.Join(gopath, "bin"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "go", "bin"), nil
}

func Test() error {
	return sh.Run("go", "test", "./...")
}

func Tidy() error {
	if err := sh.Run("go", "mod", "tidy"); err != nil {
		return err
	}
	fmt.Println("go mod tidy complete")
	return nil
}
