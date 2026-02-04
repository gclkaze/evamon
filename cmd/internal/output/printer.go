package output

type Printer interface {
	Info(msg string)
	VerboseInfo(msg string)
	Warn(msg string)
	VerboseWarn(msg string)

	Error(err error)
	ErrorWithMessage(msg string, err error)
	Success(msg string)
}
