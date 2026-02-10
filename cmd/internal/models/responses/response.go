package responses

func Ok[T any](value T) RequestResult[T] {
	return RequestResult[T]{
		Result: true,
		Value:  value,
	}
}

func OkWithMessage[T any](value T, msg string) RequestResult[T] {
	return RequestResult[T]{
		Result:  true,
		Value:   value,
		Message: msg,
	}
}

func SimpleOkMessage(msg string) RequestResult[EmptyRequestResult] {
	return RequestResult[EmptyRequestResult]{
		Result:  true,
		Message: msg,
	}
}

func Err(err error, msg string) ErrorResult {
	resp := ErrorResult{
		Result: false,
		Error:  msg,
	}

	return resp
}

func ErrWithSimpleMessage(msg string) ErrorResult {
	resp := ErrorResult{
		Result: false,
		Error:  msg,
	}

	return resp
}
