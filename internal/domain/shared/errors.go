package shared

import "errors"

// Domain errors
var (
	ErrEmptyValue      = errors.New("value cannot be empty")
	ErrInvalidLanguage = errors.New("invalid language")
	ErrInvalidType     = errors.New("invalid project type")
	ErrInvalidArch     = errors.New("invalid architecture")
)

// ErrInvalidInput crea un error de input inválido con mensaje custom
func ErrInvalidInput(msg string) error {
	return errors.New(msg)
}