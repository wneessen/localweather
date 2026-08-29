package apperror

type unexpectedError struct{}

func (*unexpectedError) Error() string {
	return "unexpected error or internal server error"
}

type invalidRequestParametersError struct{}

func (*invalidRequestParametersError) Error() string {
	return "invalid request parameters"
}

type nonPointerTargetError struct{}

func (*nonPointerTargetError) Error() string {
	return "target must be a non-nil pointer"
}

type responseBodyIsNilError struct{}

func (*responseBodyIsNilError) Error() string {
	return "response is nil"
}

var (
	ErrInvalidRequestParameters = new(invalidRequestParametersError)
	ErrNonPointerTarget         = new(nonPointerTargetError)
	ErrResponseBodyIsNil        = new(responseBodyIsNilError)
	ErrUnexpected               = new(unexpectedError)
)
