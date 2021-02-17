package error

import "fmt"

//Error custom error
type Error struct {
	Message    string `json:"message"`
	StatusCode int    `json:"code"`
}

//Errorf error
func Errorf(status int, format string, args ...interface{}) *Error {
	return &Error{StatusCode: status, Message: fmt.Sprintf(format, args...)}
}

//ClientError client error
func ClientError(format string, args ...interface{}) *Error {
	return Errorf(400, format, args...)
}

//AppError client error
func AppError(format string, args ...interface{}) *Error {
	return Errorf(500, format, args...)
}

func (e Error) String() string {
	return fmt.Sprintf("%v - [%v]", e.Message, e.StatusCode)
}
