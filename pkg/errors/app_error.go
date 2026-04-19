package errors

import "fmt"

type AppError struct {
	Code    Code
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func New(code Code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func Wrap(code Code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

func IsCode(err error, code Code) bool {
	var appErr *AppError
	if As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

func As(err error, target any) bool {
	for err != nil {
		if ae, ok := err.(*AppError); ok {
			if ptr, ok := target.(**AppError); ok {
				*ptr = ae
				return true
			}
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		err = unwrapper.Unwrap()
	}
	return false
}
