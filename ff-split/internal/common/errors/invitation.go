package errors

import "errors"

// ErrInvitationUnavailable не раскрывает различий между неизвестной, отозванной и истёкшей ссылкой.
var ErrInvitationUnavailable = errors.New("приглашение недействительно или срок его действия истёк")
