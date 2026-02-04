package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Paths struct {
	Base   string // Windows only (base root). On Linux this may be "".
	Config string
	Lib    string
	Logs   string
}

// GetConfigDir returns the system-service config directory path.
func GetConfigDir(appName string) (string, error) {
	p, err := GetPaths(appName)
	if err != nil {
		return "", err
	}
	return p.Config, nil
}

// GetLibDir returns the system-service lib/data directory path.
func GetLibDir(appName string) (string, error) {
	p, err := GetPaths(appName)
	if err != nil {
		return "", err
	}
	return p.Lib, nil
}

// GetLogsDir returns the system-service logs directory path.
func GetLogsDir(appName string) (string, error) {
	p, err := GetPaths(appName)
	if err != nil {
		return "", err
	}
	return p.Logs, nil
}

// EnsureAppDirs ensures Config/Lib/Logs dirs exist.
// Returns resolved paths.
func EnsureAppDirs(appName string) (Paths, error) {
	paths, err := GetPaths(appName)
	if err != nil {
		return Paths{}, err
	}

	for _, d := range []struct {
		name string
		path string
	}{
		{"config", paths.Config},
		{"lib", paths.Lib},
		{"logs", paths.Logs},
	} {
		if err := ensureDir(d.path); err != nil {
			return Paths{}, fmt.Errorf("ensure %s dir %q: %w", d.name, d.path, err)
		}
	}

	return paths, nil
}

// EnsurePropertiesFile ensures the app dirs exist and checks whether a properties file exists in config dir.
func EnsurePropertiesFile(appName, propertiesFileName string) (paths Paths, propsPath string, exists bool, err error) {
	if strings.TrimSpace(propertiesFileName) == "" {
		return Paths{}, "", false, fmt.Errorf("propertiesFileName is empty")
	}

	paths, err = EnsureAppDirs(appName)
	if err != nil {
		return Paths{}, "", false, err
	}

	propsPath = filepath.Join(paths.Config, propertiesFileName)
	st, statErr := os.Stat(propsPath)
	if statErr == nil {
		if st.IsDir() {
			return paths, propsPath, false, fmt.Errorf("properties path %q is a directory", propsPath)
		}
		return paths, propsPath, true, nil
	}
	if os.IsNotExist(statErr) {
		return paths, propsPath, false, nil
	}
	return paths, propsPath, false, fmt.Errorf("stat properties file %q: %w", propsPath, statErr)
}

func ensureDir(dir string) error {
	info, err := os.Stat(dir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("path exists but is not a directory")
		}
		return nil
	}
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0o755)
	}
	return err
}
