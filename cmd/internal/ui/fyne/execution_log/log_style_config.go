package executionlog

import (
	"encoding/json"
	"image/color"
	"os"
	"strconv"
	"strings"
)

// OperationStyle holds the visual parameters for one structured log type.
type OperationStyle struct {
	BorderColor string `json:"borderColor"`
	BadgeColor  string `json:"badgeColor"`
	BadgeText   string `json:"badgeText"`
}

// LogStyleConfig drives all log-line colours. Load from JSON; falls back to
// hardcoded defaults when the file is missing or malformed.
type LogStyleConfig struct {
	Styles        map[string]OperationStyle `json:"styles"`
	FreeFormColor string                    `json:"freeFormColor"`
}

// LoadLogStyleConfig reads a JSON file at path. Any error silently falls back
// to defaultLogStyleConfig so the app never crashes on a missing config.
func LoadLogStyleConfig(path string) (*LogStyleConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultLogStyleConfig(), nil
	}
	var cfg LogStyleConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultLogStyleConfig(), nil
	}
	return &cfg, nil
}

// StyleFor returns the OperationStyle for opType, or false if not configured.
func (c *LogStyleConfig) StyleFor(opType OperationType) (OperationStyle, bool) {
	s, ok := c.Styles[string(opType)]
	return s, ok
}

// FreeFormColorParsed returns the parsed free-form border color.
func (c *LogStyleConfig) FreeFormColorParsed() color.Color {
	return parseHexColor(c.FreeFormColor)
}

func defaultLogStyleConfig() *LogStyleConfig {
	return &LogStyleConfig{
		Styles: map[string]OperationStyle{
			"Label":     {BorderColor: "#4A90D9", BadgeColor: "#4A90D9", BadgeText: "Label"},
			"Operation": {BorderColor: "#E6A817", BadgeColor: "#E6A817", BadgeText: "Operation"},
			"Program":   {BorderColor: "#7B68EE", BadgeColor: "#7B68EE", BadgeText: "Program"},
		},
		FreeFormColor: "#888888",
	}
}

// parseHexColor converts a hex string (#RRGGBB or RRGGBB) to color.NRGBA.
// Returns mid-gray on any parse failure.
func parseHexColor(hex string) color.NRGBA {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return color.NRGBA{R: 128, G: 128, B: 128, A: 255}
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}
