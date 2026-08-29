package apperror

type failedToRenderErrResponseError struct{}

func (*failedToRenderErrResponseError) Error() string {
	return "failed to render error response"
}

var ErrFailedToRenderErrorResponse = new(failedToRenderErrResponseError)
