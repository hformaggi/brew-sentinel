package i18n

import (
	"errors"
	"fmt"
)

type MessageError struct {
	Source string
	Args   []any
}

func (e *MessageError) Error() string         { return fmt.Sprintf(e.Source, e.Args...) }
func Errorf(source string, args ...any) error { return &MessageError{Source: source, Args: args} }
func ErrorText(err error, tr func(string) string) string {
	if err == nil {
		return ""
	}
	var message *MessageError
	if errors.As(err, &message) {
		return fmt.Sprintf(tr(message.Source), message.Args...)
	}
	return tr(err.Error())
}
