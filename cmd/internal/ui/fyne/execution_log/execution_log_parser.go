package executionlog

import "strings"

// OperationType classifies a structured log line for styling.
// The string value matches the key in LogStyleConfig.Styles.
type OperationType string

const (
	OperationTypeLabel     OperationType = "Label"
	OperationTypeOperation OperationType = "Operation"
	OperationTypeProgram   OperationType = "Program"
)

// ParsedLine is the result of inspecting a raw log message string.
type ParsedLine struct {
	IsStructured  bool
	OperationType OperationType
}

// ParseLogLine inspects the message text and returns its structural classification.
// Free-form lines (no recognised prefix) have IsStructured = false.
func ParseLogLine(msg string) ParsedLine {
	low := strings.ToLower(strings.TrimSpace(msg))
	switch {
	case strings.HasPrefix(low, "operation:") || strings.HasPrefix(low, "[operation]"):
		return ParsedLine{IsStructured: true, OperationType: OperationTypeOperation}
	case strings.HasPrefix(low, "label:") || strings.HasPrefix(low, "[label]"):
		return ParsedLine{IsStructured: true, OperationType: OperationTypeLabel}
	case strings.HasPrefix(low, "program:") || strings.HasPrefix(low, "[program]"):
		return ParsedLine{IsStructured: true, OperationType: OperationTypeProgram}
	default:
		return ParsedLine{IsStructured: false}
	}
}
