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
		{"geobus: ErrNoGeobusProvider", ErrNoGeobusProvider, wantType[*noGeobusEnabledError]},
		{"geobus: ErrNoValidCoordinates", ErrNoValidCoordinates, wantType[*noValidCoordinatesFoundError]},
		{"geocoder: ErrGeoCoderRequired", ErrGeoCoderRequired, wantType[*geoCoderRequiredError]},
		{"http: ErrUnexpected", ErrUnexpected, wantType[*unexpectedError]},
		{"http: ErrNonPointerTarget", ErrNonPointerTarget, wantType[*nonPointerTargetError]},
		{"http: ErrInvalidRequestParameters", ErrInvalidRequestParameters, wantType[*invalidRequestParametersError]},
		{"http: ErrResponseBodyIsNil", ErrResponseBodyIsNil, wantType[*responseBodyIsNilError]},
		{"location: ErrLocationNotSet", ErrLocationNotSet, wantType[*locationNotSetError]},
		{"logger: ErrLoggerRequired", ErrLoggerRequired, wantType[*loggerRequiredError]},
		{
			"render: ErrFailedToRenderErrorResponse", ErrFailedToRenderErrorResponse,
			wantType[*failedToRenderErrResponseError],
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
