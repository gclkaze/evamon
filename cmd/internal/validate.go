package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
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

const (
	MaxArgsCount = 20
	MaxTagsCount = 20
	MaxArgLen    = 256
	MaxTagLen    = 64
	MaxDescLen   = 200
)

// ParseCSV splits a comma-separated string into a trimmed slice.
// It drops empty entries (",,").
// If input is empty/whitespace => returns nil, nil.

func ParseCSV(input string) []string {
	s := strings.TrimSpace(input)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func ValidateArgs(args []string) error {
	if len(args) == 0 {
		return nil
	}
	if len(args) > MaxArgsCount {
		return fmt.Errorf("args has too many items (max %d)", MaxArgsCount)
	}
	for i, a := range args {
		a = strings.TrimSpace(a)
		if a == "" {
			return fmt.Errorf("args[%d] is empty", i)
		}
		if utf8.RuneCountInString(a) > MaxArgLen {
			return fmt.Errorf("args[%d] too long (max %d chars)", i, MaxArgLen)
		}
	}
	return nil
}

func ValidateTags(tags []string) error {
	if len(tags) == 0 {
		return nil
	}
	if len(tags) > MaxTagsCount {
		return fmt.Errorf("tags has too many items (max %d)", MaxTagsCount)
	}
	for i, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" {
			return fmt.Errorf("tags[%d] is empty", i)
		}
		if utf8.RuneCountInString(t) > MaxTagLen {
			return fmt.Errorf("tags[%d] too long (max %d chars)", i, MaxTagLen)
		}
		// Optional: enforce no spaces inside a tag (common convention)
		// if strings.ContainsAny(t, " \t") { return fmt.Errorf("tags[%d] contains whitespace", i) }
	}
	return nil
}

func ValidateDescription(desc string) error {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return nil // description is optional
	}
	if utf8.RuneCountInString(desc) > MaxDescLen {
		return fmt.Errorf("description too long (max %d chars)", MaxDescLen)
	}
	return nil
}
