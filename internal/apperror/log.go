package apperror

type loggerRequiredError struct{}

func (*loggerRequiredError) Error() string {
	return "function requires a logger"
}

var ErrLoggerRequired = new(loggerRequiredError)
