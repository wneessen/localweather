package apperror

type CurrentWeatherdataNotFoundError struct{}

func (*CurrentWeatherdataNotFoundError) Error() string {
	return "no current weather data found"
}

var ErrCurrentWeatherdataNotFound = new(CurrentWeatherdataNotFoundError)
