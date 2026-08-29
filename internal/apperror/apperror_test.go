package apperror

import (
	"errors"
	"testing"
)

func TestApperrors(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		match func(error) bool
	}{
		{"geobus: ErrNoGeobusProvider", ErrNoGeobusProvider, wantType[*NoGeobusEnabledError]},
		{"geobus: ErrNoValidCoordinates", ErrNoValidCoordinates, wantType[*NoValidCoordinatesFoundError]},
		{"geocoder: ErrGeoCoderRequired", ErrGeoCoderRequired, wantType[*GeoCoderRequiredError]},
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
			if !test.match(test.err) {
				t.Errorf("error did not match expected type: %T / %s", test.err, test.err)
			}
		})
	}
}

func wantType[E error](err error) bool {
	_, ok := errors.AsType[E](err)
	return ok
}
