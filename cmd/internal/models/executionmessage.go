// {"ltype":"info","msg":"Line:14 MetricsIndexSave Operation. Message: Saving Metrics Index :soilcrates-up","t":1771158300810}
package models

type ExecutionMessage struct {
	LType string `json:"ltype"`
	Msg   string `json:"msg"`
	T     int64  `json:"t"`
}
