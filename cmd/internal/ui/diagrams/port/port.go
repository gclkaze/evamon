// internal/ui/diagrams/port/port.go
package port

import (
	"time"
)

type DataPoint[T any] struct {
	At    time.Time
	Value T
}
