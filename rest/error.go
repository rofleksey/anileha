package rest

import (
	"net/http"
)

type StatusError struct {
	StatusCode int
	Message    string
}

func (e *StatusError) Error() string {
	return e.Message
}

func WrapError(statusCode int, err error) *StatusError {
	return &StatusError{
		StatusCode: statusCode,
		Message:    err.Error(),
	}
}

func ErrNotFound(msg string) *StatusError {
	return &StatusError{
		StatusCode: http.StatusNotFound,
		Message:    msg,
	}
}

func ErrBadRequest(msg string) *StatusError {
	return &StatusError{
		StatusCode: http.StatusBadRequest,
		Message:    msg,
	}
}

func ErrUnauthorized(msg string) *StatusError {
	return &StatusError{
		StatusCode: http.StatusUnauthorized,
		Message:    msg,
	}
}

func ErrForbidden(msg string) *StatusError {
	return &StatusError{
		StatusCode: http.StatusForbidden,
		Message:    msg,
	}
}

func ErrInternal(msg string) *StatusError {
	return &StatusError{
		StatusCode: http.StatusInternalServerError,
		Message:    msg,
	}
}

var (
	ErrNotFoundInst                   = ErrNotFound("not found")
	ErrReadyFileNotFound              = ErrNotFound("ready file not found")
	ErrCreationFailed                 = ErrInternal("creation failed")
	ErrDeletionFailed                 = ErrInternal("deletion failed")
	ErrInvalidIndicesString           = ErrBadRequest("invalid indices string")
	ErrFileIsNotReadyToBeConverted    = ErrBadRequest("file is not ready to be converted")
	ErrAlreadyStarted                 = ErrBadRequest("already started")
	ErrAlreadyStopped                 = ErrBadRequest("already stopped")
	ErrUnauthorizedInst               = ErrUnauthorized("unauthorized")
	ErrLinkExpired                    = ErrBadRequest("link expired")
	ErrInvalidPassword                = ErrUnauthorized("invalid password")
	ErrUserWithThisEmailAlreadyExists = ErrBadRequest("user with this email already exists")
	ErrUserWithThisLoginAlreadyExists = ErrBadRequest("user with this login already exists")
	ErrSessionSavingFailed            = ErrInternal("session saving failed")
)
