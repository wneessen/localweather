package apperror

type UnexpectedError struct{}

func (*UnexpectedError) Error() string {
	return "unexpected error or internal server error"
}

type InvalidRequestParametersError struct{}

func (*InvalidRequestParametersError) Error() string {
	return "invalid request parameters"
}

type NonPointerTargetError struct{}

func (*NonPointerTargetError) Error() string {
	return "target must be a non-nil pointer"
}

type ResponseBodyIsNilError struct{}

func (*ResponseBodyIsNilError) Error() string {
	return "response is nil"
}

type HTTPClientRequiredError struct{}

func (*HTTPClientRequiredError) Error() string {
	return "http client is required"
}

var (
	ErrHTTPClientRequired       = new(HTTPClientRequiredError)
	ErrInvalidRequestParameters = new(InvalidRequestParametersError)
	ErrNonPointerTarget         = new(NonPointerTargetError)
	ErrResponseBodyIsNil        = new(ResponseBodyIsNilError)
	ErrUnexpected               = new(UnexpectedError)
)
