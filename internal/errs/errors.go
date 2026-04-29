package errs

import (
	"errors"
	"fmt"
)

type Code string

const (
	CodeValidation Code = "VALIDATION"
	CodeConfig     Code = "CONFIG"
	CodeConnect    Code = "CONNECT"
	CodeQuery      Code = "QUERY"
	CodeInternal   Code = "INTERNAL"
)

type Error struct {
	Code Code
	Op   string
	Msg  string
	Err  error
}

func (e *Error) Error() string {
	switch {
	case e.Msg != "" && e.Op != "":
		return fmt.Sprintf("%s: %s", e.Op, e.Msg)
	case e.Msg != "":
		return e.Msg
	case e.Op != "":
		return e.Op
	default:
		return "unknown error"
	}
}

func (e *Error) Unwrap() error { return e.Err }

func E(code Code, op, msg string, err error) error {
	return &Error{
		Code: code,
		Op:   op,
		Msg:  msg,
		Err:  err,
	}
}

func Validation(op, msg string) error {
	return E(CodeValidation, op, msg, nil)
}

func Message(err error) string {
	if err == nil {
		return ""
	}
	var e *Error
	if errors.As(err, &e) && e.Msg != "" {
		return e.Msg
	}
	return err.Error()
}
