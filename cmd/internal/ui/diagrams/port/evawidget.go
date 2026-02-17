package port

import "time"

type EvaWidget interface {
	Push(at time.Time, val any)
}
