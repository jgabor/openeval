package doctor

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func WriteJSON(w io.Writer, report Report) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func WriteHuman(w io.Writer, report Report) {
	_, _ = fmt.Fprintf(w, "OpenEval doctor (%s): %s\n", report.Agent, report.Status)
	for _, check := range report.Checks {
		_, _ = fmt.Fprintf(w, "%s %-15s %s\n", strings.ToUpper(string(check.Status)), check.ID, check.Summary)
		if check.Remediation != "" && check.Status != StatusPass {
			_, _ = fmt.Fprintf(w, "  fix: %s\n", check.Remediation)
		}
	}
}
