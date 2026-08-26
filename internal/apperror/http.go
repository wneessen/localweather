package apperror

type UnexpectedError struct{}

func (*UnexpectedError) Error() string {
	return "unexpected error or internal server error"
}

type InvalidRequestParametersError struct{}

func (*InvalidRequestParametersError) Error() string {
	return "invalid request parameters"
}

var (
	ErrInvalidRequestParameters = new(InvalidRequestParametersError)
	ErrUnexpected               = new(UnexpectedError)
)
