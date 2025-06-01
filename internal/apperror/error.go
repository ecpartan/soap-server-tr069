package apperror

import (
	"encoding/json"
	"fmt"
)

type AppError struct {
	Err        error  `json:"-"`
	Message    string `json:"message"`
	DevMessage string `json:"dev_message"`
	Code       string `json:"code"`
}

var (
	ErrNotFound     = NewAppError("Not Found", "Not Found", "404")
	ErrAlreadyExist = NewAppError("Already Exist", "Already Exist", "409")
)

func (e *AppError) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) Marshal() []byte {
	marshal, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return marshal
}

func NewAppError(message, devMessage, code string) *AppError {
	return &AppError{
		Err:        fmt.Errorf("%s", message),
		Message:    message,
		DevMessage: devMessage,
		Code:       code,
	}
}
func systemError(developerMessage string) *AppError {
	return NewAppError("system error", "TS-000001", developerMessage)
}
