package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/gclkaze/evamon/pkg/utils"
)

const (
	MaxArgsCount = 20
	MaxTagsCount = 20
	MaxArgLen    = 256
	MaxTagLen    = 64
	MaxDescLen   = 200
)

type JobAddRequest struct {
	Schedule     string   `json:"schedule"`
	ScriptPath   string   `json:"scriptPath"`
	Args         []string `json:"args,omitempty"`
	Description  string   `json:"description,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	AllowOverlap bool     `json:"allowOverlap,omitempty"`
}

type JobUpdateRequest struct {
	ID           string   `json:"id"`
	Schedule     string   `json:"schedule"`
	Script       string   `json:"script"`
	Args         []string `json:"args,omitempty"`
	Description  string   `json:"description,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	AllowOverlap bool     `json:"allowOverlap,omitempty"`
}

func NewJobUpdateRequest(ID string, schedule string, script string, args []string, description string, tags []string, allowedOverlap bool) *JobUpdateRequest {
	return &JobUpdateRequest{ID: ID, Schedule: schedule, Script: script, Args: args, Description: description, Tags: tags, AllowOverlap: allowedOverlap}
}

// NewJobAddRequest performs ALL preprocessing:
// - trims schedule/script/description
// - normalizes args/tags (trim + drop empty)
// - converts Script to absolute cleaned path when possible
func NewJobAddRequest(schedule string, script string, args []string, description string, tags []string, allowedOverlap bool) *JobAddRequest {
	schedule = strings.TrimSpace(schedule)
	script = strings.TrimSpace(script)
	description = strings.TrimSpace(description)

	// Normalize slices: trim entries + drop empty
	args = utils.NormalizeSlice(args)
	tags = utils.NormalizeSlice(tags)

	// Normalize script path to absolute if it isn't empty
	if script != "" {
		if abs, err := filepath.Abs(script); err == nil {
			script = filepath.Clean(abs)
		} else {
			// If Abs fails, keep original; IsValid will still fail existence checks.
			script = filepath.Clean(script)
		}
	}

	return &JobAddRequest{
		Schedule:     schedule,
		ScriptPath:   script,
		Args:         args,
		Description:  description,
		Tags:         tags,
		AllowOverlap: allowedOverlap,
	}
}

// IsValid performs ALL validation:
// - schedule/script required
// - script must exist and be a file
// - description max length
// - args/tags max count and max item length, no empty items
func (r *JobAddRequest) IsValid() error {
	if r == nil {
		return fmt.Errorf("job add request is nil")
	}

	if r.Schedule == "" {
		return fmt.Errorf("schedule is empty")
	}
	if r.ScriptPath == "" {
		return fmt.Errorf("script is empty")
	}

	// Script must exist and be a file
	info, err := os.Stat(r.ScriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("script does not exist: %s", r.ScriptPath)
		}
		return fmt.Errorf("cannot access script %q: %w", r.ScriptPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("script must be a file, got directory: %s", r.ScriptPath)
	}

	// Description optional
	if r.Description != "" && utf8.RuneCountInString(r.Description) > MaxDescLen {
		return fmt.Errorf("description too long (max %d chars)", MaxDescLen)
	}

	// Args optional
	if len(r.Args) > MaxArgsCount {
		return fmt.Errorf("args has too many items (max %d)", MaxArgsCount)
	}
	for i, a := range r.Args {
		a = strings.TrimSpace(a)
		if a == "" {
			return fmt.Errorf("args[%d] is empty", i)
		}
		if utf8.RuneCountInString(a) > MaxArgLen {
			return fmt.Errorf("args[%d] too long (max %d chars)", i, MaxArgLen)
		}
	}

	// Tags optional
	if len(r.Tags) > MaxTagsCount {
		return fmt.Errorf("tags has too many items (max %d)", MaxTagsCount)
	}
	for i, t := range r.Tags {
		t = strings.TrimSpace(t)
		if t == "" {
			return fmt.Errorf("tags[%d] is empty", i)
		}
		if utf8.RuneCountInString(t) > MaxTagLen {
			return fmt.Errorf("tags[%d] too long (max %d chars)", i, MaxTagLen)
		}
	}

	return nil
}
