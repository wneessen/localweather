package apperror

type locationNotSetError struct{}

func (*locationNotSetError) Error() string {
	return "no current location set"
}

var ErrLocationNotSet = new(locationNotSetError)
