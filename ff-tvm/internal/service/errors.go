package service

import "errors"

var (
	ErrAccessDenied     = errors.New("access denied")
	ErrTicketExpired    = errors.New("ticket expired")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrServiceNotFound  = errors.New("service not found")
	ErrInvalidSecret    = errors.New("invalid secret")
)
