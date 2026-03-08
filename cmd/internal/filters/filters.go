package filters

type CombineMode int

const (
	CombineAND CombineMode = iota
	CombineOR
)

type FilterSet struct {
	Expressions []string
	Mode        CombineMode
}
