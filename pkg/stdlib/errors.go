package stdlib

import "fmt"

type ArgumentError struct {
	name    string
	wrapped error
}

func NewArgumentError(name string, wrapped error) ArgumentError {
	return ArgumentError{name: name, wrapped: wrapped}
}

func (e ArgumentError) Error() string {
	return fmt.Sprintf("%s(): argument error: %s", e.name, e.wrapped)
}

func (e ArgumentError) Unwrap() error {
	return e.wrapped
}

type InvalidNumberOfArgumentsError struct {
	Expected int
	Actual   int
	Message  string
}

func NewInvalidNumberOfArgumentsError(name, message string, expected, actual int) error {
	return NewArgumentError(name, InvalidNumberOfArgumentsError{Message: message, Actual: actual, Expected: expected})
}

func (e InvalidNumberOfArgumentsError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("accepts exactly %d argument, %d provided", e.Expected, e.Actual)
	}

	return fmt.Sprintf(e.Message, e.Expected, e.Actual)
}

type InvalidArgumentTypeError struct {
	Message string
}

func (e InvalidArgumentTypeError) Error() string {
	return e.Message
}

func NewInvalidArgumentTypeError(name, message string) error {
	return NewArgumentError(name, InvalidArgumentTypeError{Message: message})
}
