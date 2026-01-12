package script

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
)

var placeholderRegex = regexp.MustCompile(`\{\{(.*?)\}\}`)

// GetPlaceholders returns a list of placeholder names found in the command string.
// e.g., "git commit -m {{message}}" returns ["message"]
func GetPlaceholders(cmd string) []string {
	matches := placeholderRegex.FindAllStringSubmatch(cmd, -1)
	var keys []string
	seen := make(map[string]bool)
	for _, m := range matches {
		if len(m) > 1 && !seen[m[1]] {
			keys = append(keys, m[1])
			seen[m[1]] = true
		}
	}
	return keys
}

// ReplacePlaceholders replaces {{key}} with values from the map.
func ReplacePlaceholders(cmd string, values map[string]string) string {
	return placeholderRegex.ReplaceAllStringFunc(cmd, func(s string) string {
		key := s[2 : len(s)-2] // remove {{ and }}
		if val, ok := values[key]; ok {
			return val
		}
		return s
	})
}

// Run executes the command in a new terminal window.
func Run(cmdStr string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		// Spawns a new PowerShell window.
		// We append Read-Host to ensure the window stays open so the user can see the output.
		// The title is set to "Zenith Script".
		wrappedCmd := fmt.Sprintf(`%s; Write-Host -ForegroundColor Green "`+"`n"+`[Process completed]"; Read-Host "Press Enter to exit..."`, cmdStr)
		cmd = exec.Command("cmd", "/c", "start", "Zenith Script", "powershell", "-NoProfile", "-Command", wrappedCmd)
	} else {
		// Fallback for Linux/Mac (simplified, assumes default terminal availability)
		// For a more robust solution, we'd check for xterm, gnome-terminal, etc.
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	// Start the command and immediately return, detaching it from the TUI.
	return cmd.Start()
}
