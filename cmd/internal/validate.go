package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FileExists returns nil if path exists (file or directory), otherwise an error.
// If you want to require "file only", use RequireFileExists.
func FileExists(path string) error {
	p := strings.TrimSpace(path)
	if p == "" {
		return fmt.Errorf("path is empty")
	}
	_, err := os.Stat(p)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", p)
	}
	return fmt.Errorf("cannot access path %q: %w", p, err)
}

// RequireFileExists returns nil only if path exists AND is a file (not a directory).
func RequireFileExists(path string) error {
	p := strings.TrimSpace(path)
	if p == "" {
		return fmt.Errorf("path is empty")
	}
	info, err := os.Stat(p)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("expected a file but got a directory: %s", p)
		}
		return nil
	}
	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", p)
	}
	return fmt.Errorf("cannot access file %q: %w", p, err)
}

// RequireAbsPath returns an absolute, cleaned version of the path, or an error.
func RequireAbsPath(path string) (string, error) {
	p := strings.TrimSpace(path)
	if p == "" {
		return "", fmt.Errorf("path is empty")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("cannot make absolute path from %q: %w", p, err)
	}
	return filepath.Clean(abs), nil
}

// NonEmpty ensures a string is not empty/whitespace.
func NonEmpty(name, val string) error {
	if strings.TrimSpace(val) == "" {
		return fmt.Errorf("%s is empty", name)
	}
	return nil
}

// IsValidId checks job-id length == 10 and a safe charset.
// If you only want length, delete the regex part.
func IsValidId(id string) error {
	id = strings.TrimSpace(id)
	if len(id) != 10 {
		return fmt.Errorf("job-id must be exactly 10 characters (got %d)", len(id))
	}

	// Safe charset: letters, numbers, underscore, hyphen
	var re = regexp.MustCompile(`^[A-Za-z0-9_-]{10}$`)
	if !re.MatchString(id) {
		return fmt.Errorf("job-id contains invalid characters (allowed: A-Z a-z 0-9 _ -)")
	}
	return nil
}

// IntervalString currently just checks non-empty.
// You can later replace this with a cron parser validation.
func IntervalString(interval string) error {
	return NonEmpty("interval-string", interval)
}
