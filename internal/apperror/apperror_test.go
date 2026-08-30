package apperror

import (
	"errors"
	"testing"
)

func TestApperrors(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		match func(error) (error, bool)
	}{
		{"coordinates: CoordinateParsingError", &CoordinateParsingError{}, wantType[*CoordinateParsingError]},
		{
			"coordinates: NoCoordinatesFoundForAddressError", &NoCoordinatesFoundForAddressError{},
			wantType[*NoCoordinatesFoundForAddressError],
		},
		{"geobus: ErrNoGeobusProvider", ErrNoGeobusProvider, wantType[*NoGeobusEnabledError]},
		{"geobus: ErrNoValidCoordinates", ErrNoValidCoordinates, wantType[*NoValidCoordinatesFoundError]},
		{"geocoder: ErrGeoCoderRequired", ErrGeoCoderRequired, wantType[*GeoCoderRequiredError]},
		{"location: GeoLocationAPIFetchError", &GeoLocationAPIFetchError{}, wantType[*GeoLocationAPIFetchError]},
		{"location: ErrCurrentLocationNotSet", ErrCurrentLocationNotSet, wantType[*CurrentLocationNotSetError]},
		{"logger: ErrLoggerRequired", ErrLoggerRequired, wantType[*LoggerRequiredError]},
		{"http: ErrHTTPClientRequired", ErrHTTPClientRequired, wantType[*HTTPClientRequiredError]},
		{"http: ErrInvalidRequestParameters", ErrInvalidRequestParameters, wantType[*InvalidRequestParametersError]},
		{"http: ErrNonPointerTarget", ErrNonPointerTarget, wantType[*NonPointerTargetError]},
		{"http: ErrResponseBodyIsNil", ErrResponseBodyIsNil, wantType[*ResponseBodyIsNilError]},
		{"http: ErrUnexpected", ErrUnexpected, wantType[*UnexpectedError]},
		{
			"render: ErrFailedToRenderErrorResponse", ErrFailedToRenderErrorResponse,
			wantType[*FailedToRenderErrResponseError],
		},
		{
			"weather: ErrCurrentWeatherdataNotFound", ErrCurrentWeatherdataNotFound,
			wantType[*CurrentWeatherdataNotFoundError],
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualErr, ok := test.match(test.err)
			if !ok {
				t.Errorf("error did not match expected type: %T (%s), got: %T (%s)", test.err, test.err,
					actualErr, actualErr)
			}
			if got := actualErr.Error(); got != test.err.Error() {
				t.Errorf("Error() = %q, want %q", got, test.err.Error())
			}
		})
	}
}

func wantType[E error](err error) (error, bool) {
	actualErr, ok := errors.AsType[E](err)
	return actualErr, ok
}
