package methods

import "github.com/ecpartan/soap-server-tr069/pkg/errors"

type MethodsError struct {
	Err     error  `json:"-"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

func NewAppError(code, message string) *MethodsError {
	return &MethodsError{
		Err:     errors.New(message),
		Code:    code,
		Message: message,
	}
}

func (e *MethodsError) Error() string {
	return e.Err.Error()
}

func (e *MethodsError) Unwrap() error { return e.Err }

var ErrNotFound = errors.New("not found")
var ErrNotfoundSerial = errors.New("not found serial")
var ErrNotfoundTree = errors.New("not found tree")
var ErrNotfoundParams = errors.New("not found params")
