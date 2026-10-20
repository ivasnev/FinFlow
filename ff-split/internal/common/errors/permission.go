package errors

import "errors"

// ErrForbidden означает отсутствие права на изменение конкретной записи.
var ErrForbidden = errors.New("недостаточно прав для этого действия")
