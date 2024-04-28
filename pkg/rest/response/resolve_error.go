package response

import (
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidCredentials = "invalid credentials"
	ErrUserExist          = "user already exists"
	ErrUnauthorized       = "Unauthenticated"
)

func ResolveError(err error) Error {
	st, ok := status.FromError(err)
	if !ok {
		return NewInternalError()
	}

	switch st.Message() {
	case ErrUserExist:
		return NewUserExistError()
	case ErrInvalidCredentials:
		return NewInvalidCredentialsError()
	case ErrUnauthorized:
		return NewUnauthorizedError()
	default:
		return NewInternalError()
	}
}
