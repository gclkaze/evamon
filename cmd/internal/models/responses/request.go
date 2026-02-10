package responses

import "encoding/json"

type RequestResult[T any] struct {
	Result  bool   `json:"result"`
	Value   T      `json:"value,omitempty"`
	Message string `json:"message,omitempty"`
}
type EmptyRequestResult struct{}
type ErrorResult struct {
	Result  bool   `json:"result"`
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

func ToBytes[T any](r RequestResult[T]) ([]byte, error) {
	return json.Marshal(r)
}
