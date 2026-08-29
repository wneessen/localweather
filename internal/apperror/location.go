package apperror

type LocationNotSetError struct{}

func (*LocationNotSetError) Error() string {
	return "no current location set"
}

var ErrLocationNotSet = new(LocationNotSetError)
