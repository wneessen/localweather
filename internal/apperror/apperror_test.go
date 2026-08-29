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
		{"geobus: ErrNoGeobusProvider", ErrNoGeobusProvider, wantType[*NoGeobusEnabledError]},
		{"geobus: ErrNoValidCoordinates", ErrNoValidCoordinates, wantType[*NoValidCoordinatesFoundError]},
		{"geocoder: ErrGeoCoderRequired", ErrGeoCoderRequired, wantType[*GeoCoderRequiredError]},
		{"coordinates: CoordinateParsingError", &CoordinateParsingError{}, wantType[*CoordinateParsingError]},
		{
			"coordinates: NoCoordinatesFoundForAddressError", &NoCoordinatesFoundForAddressError{},
			wantType[*NoCoordinatesFoundForAddressError],
		},
		{"location: GeoLocationAPIFetchError", &GeoLocationAPIFetchError{}, wantType[*GeoLocationAPIFetchError]},
		{"http: ErrHTTPClientRequired", ErrHTTPClientRequired, wantType[*HTTPClientRequiredError]},
		{"http: ErrInvalidRequestParameters", ErrInvalidRequestParameters, wantType[*InvalidRequestParametersError]},
		{"http: ErrNonPointerTarget", ErrNonPointerTarget, wantType[*NonPointerTargetError]},
		{"http: ErrResponseBodyIsNil", ErrResponseBodyIsNil, wantType[*ResponseBodyIsNilError]},
		{"http: ErrUnexpected", ErrUnexpected, wantType[*UnexpectedError]},
		{"location: ErrLocationNotSet", ErrLocationNotSet, wantType[*LocationNotSetError]},
		{"logger: ErrLoggerRequired", ErrLoggerRequired, wantType[*LoggerRequiredError]},
		{
			"render: ErrFailedToRenderErrorResponse", ErrFailedToRenderErrorResponse,
			wantType[*FailedToRenderErrResponseError],
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
