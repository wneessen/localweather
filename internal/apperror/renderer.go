package apperror

type FailedToRenderErrorResponseError struct{}

func (*FailedToRenderErrorResponseError) Error() string {
	return "failed to render error response"
}

var ErrFailedToRenderErrorResponse = new(FailedToRenderErrorResponseError)
