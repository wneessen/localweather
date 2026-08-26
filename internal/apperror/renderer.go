package apperror

type FailedToRenderErrResponseError struct{}

func (*FailedToRenderErrResponseError) Error() string {
	return "failed to render error response"
}

var ErrFailedToRenderErrorResponse = new(FailedToRenderErrResponseError)
