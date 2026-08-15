package examples

import "embed"

// Files contains the example assets needed by an installed OpenEval binary.
//
//go:embed scenarios/example-fixtures scenarios/example-fixtures/fixtures/maintainer_tools/__init__.py skills/demo-skill
var Files embed.FS
