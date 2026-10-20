package errors

import "errors"

// ErrEventArchived запрещает изменение данных до явного восстановления события.
var ErrEventArchived = errors.New("событие в архиве; восстановите его, чтобы вносить изменения")
