package apperror

type UnexpectedError struct{}

func (*UnexpectedError) Error() string {
	return "unexpected error or internal server error"
}

type AuthTokenInvalidError struct{}

func (*AuthTokenInvalidError) Error() string {
	return "authentication token missing, invalid or expired"
}

type UserNotAuthorizedError struct{}

func (*UserNotAuthorizedError) Error() string {
	return "user not authorized"
}

type InvalidUserTypeError struct{}

func (*InvalidUserTypeError) Error() string {
	return "invalid user type"
}

type InvalidRequestParametersError struct{}

func (*InvalidRequestParametersError) Error() string {
	return "invalid request parameters"
}

var (
	ErrInvalidAuthToken         = new(AuthTokenInvalidError)
	ErrInvalidRequestParameters = new(InvalidRequestParametersError)
	ErrUserNotAuthorized        = new(UserNotAuthorizedError)
	ErrUserTypeNotValid         = new(InvalidUserTypeError)
	ErrUnexpected               = new(UnexpectedError)
)
