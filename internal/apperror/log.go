package apperror

type LoggerRequiredError struct{}

func (*LoggerRequiredError) Error() string {
	return "function requires a logger"
}

var ErrLoggerRequired = new(LoggerRequiredError)
