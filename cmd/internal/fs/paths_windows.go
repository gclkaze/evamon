//go:build windows

package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetPaths returns the Windows layout:
//
// Base:   %ProgramFiles%\Evacron
// Config: %ProgramFiles%\Evacron\config
// Lib:    %ProgramFiles%\Evacron\lib
// Logs:   %ProgramFiles%\Evacron\logs
func GetPaths(appName string) (Paths, error) {
	appName = strings.TrimSpace(appName)
	if appName == "" {
		return Paths{}, fmt.Errorf("appName is empty")
	}

	programFiles := os.Getenv("ProgramData")
	if programFiles == "" {
		// Fallback if env isn't present (rare)
		return Paths{}, fmt.Errorf("ProgramData env var is empty")
	}

	base := filepath.Join(programFiles, appName)
	return Paths{
		Base:   base,
		Config: filepath.Join(base, "config"),
		Lib:    filepath.Join(base, "lib"),
		Logs:   filepath.Join(base, "logs"),
	}, nil
}
