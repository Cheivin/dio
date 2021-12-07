package errors

import (
	"errors"
	"net/http"
)

type Error struct {
	error
	Code int
	Msg  string
}

var NoRoute = ErrMsg(http.StatusNotFound, "Not Found")
var NoMethod = ErrMsg(http.StatusMethodNotAllowed, "Method Not Allowed")

func (e *Error) Error() string {
	if e.Msg == "" {
		return e.error.Error()
	}
	return e.Msg
}

func (e *Error) Cause() error {
	return e.error
}

func (e *Error) CauseBy(cause error) Error {
	copyE := *e
	copyE.error = cause
	return copyE
}

func Err(code int, msg string, cause error) Error {
	return Error{
		error: cause,
		Code:  code,
		Msg:   msg,
	}
}

func ErrMsg(code int, msg string) Error {
	return Error{
		error: errors.New(msg),
		Code:  code,
		Msg:   msg,
	}
}

func ErrWith(err Error, cause error) Error {
	return Error{
		error: cause,
		Code:  err.Code,
		Msg:   err.Msg,
	}
}
