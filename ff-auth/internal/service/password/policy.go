package password

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// ErrInvalid позволяет обработчику отличить нарушение политики пароля от сбоя хранилища.
var ErrInvalid = errors.New("некорректный пароль")

// Validate проверяет длину в символах и предел bcrypt в байтах до вычисления хеша.
func Validate(value string, minimum int) error {
	if minimum <= 0 {
		minimum = 8
	}
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) < minimum {
		return fmt.Errorf("%w: требуется минимум %d символов", ErrInvalid, minimum)
	}
	if len(value) > 72 {
		return fmt.Errorf("%w: допускается не более 72 байт", ErrInvalid)
	}
	return nil
}
