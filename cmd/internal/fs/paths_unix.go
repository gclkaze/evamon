//go:build !windows

package fs

import (
	"fmt"
	"path/filepath"
	"strings"
)

// GetPaths returns the Linux layout:
//
// Config: /etc/evacron
// Lib:    /var/lib/evacron
// Logs:   /var/log/evacron
func GetPaths(appName string) (Paths, error) {
	appName = strings.TrimSpace(appName)
	if appName == "" {
		return Paths{}, fmt.Errorf("appName is empty")
	}

	// Linux convention: lower-case folder name
	name := strings.ToLower(appName)

	return Paths{
		Base:   "",
		Config: filepath.Join(string(filepath.Separator), "etc", name),
		Lib:    filepath.Join(string(filepath.Separator), "var", "lib", name),
		Logs:   filepath.Join(string(filepath.Separator), "var", "log", name),
	}, nil
}
