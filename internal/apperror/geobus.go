package apperror

type noGeobusEnabledError struct{}

func (*noGeobusEnabledError) Error() string {
	return "no geobus provider enabled"
}

type noValidCoordinatesFoundError struct{}

func (*noValidCoordinatesFoundError) Error() string {
	return "no valid coordinates found in coordinates file"
}

var (
	ErrNoGeobusProvider   = new(noGeobusEnabledError)
	ErrNoValidCoordinates = new(noValidCoordinatesFoundError)
)
