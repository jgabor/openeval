package instrument

import (
	"encoding/json"
	"fmt"
	"strings"
)

const openEvalHookMarker = "hook --agent cursor"

var openEvalHookEvents = []string{
	"sessionStart",
	"sessionEnd",
	"preToolUse",
	"postToolUse",
}

type hooksFile struct {
	Version int                    `json:"version"`
	Hooks   map[string][]hookEntry `json:"hooks"`
}

type hookEntry struct {
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
	Matcher string `json:"matcher,omitempty"`
}

func mergeOpenEvalHooks(existing []byte, openevalCmd string) ([]byte, error) {
	var doc hooksFile
	if len(strings.TrimSpace(string(existing))) > 0 {
		if err := json.Unmarshal(existing, &doc); err != nil {
			return nil, fmt.Errorf("parse existing hooks.json: %w", err)
		}
	}
	if doc.Version == 0 {
		doc.Version = 1
	}
	if doc.Hooks == nil {
		doc.Hooks = map[string][]hookEntry{}
	}
	entry := hookEntry{Command: openevalCmd, Timeout: 15}
	for _, event := range openEvalHookEvents {
		if hasOpenEvalHook(doc.Hooks[event]) {
			continue
		}
		doc.Hooks[event] = append(doc.Hooks[event], entry)
	}
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

func hasOpenEvalHook(entries []hookEntry) bool {
	for _, e := range entries {
		if strings.Contains(e.Command, openEvalHookMarker) {
			return true
		}
	}
	return false
}
