package errors

import "fmt"

type EntityNotFoundError struct {
	ID         string
	EntityName string
}

func NewEntityNotFoundError(id string, entityName string) *EntityNotFoundError {
	return &EntityNotFoundError{
		ID:         id,
		EntityName: entityName,
	}
}

func (e *EntityNotFoundError) Error() string {
	return fmt.Sprintf("entity %s with id %s not found", e.EntityName, e.ID)
}

type ValidationError struct {
	Field   string
	Message string
}

func NewValidationError(field string, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

type AlreadyExistsError struct {
	EntityName string
	Message    string
}

func NewAlreadyExistsError(entityName string, message string) *AlreadyExistsError {
	return &AlreadyExistsError{
		EntityName: entityName,
		Message:    message,
	}
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("Object %s already exists: %s", e.EntityName, e.Message)
}

type LogicError struct {
	Message string
}

func NewLogicError(message string) *LogicError {
	return &LogicError{
		Message: message,
	}
}

func (e *LogicError) Error() string {
	return fmt.Sprintf("Logic error: %s", e.Message)
}
