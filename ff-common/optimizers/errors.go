package optimizers

import "fmt"

// ValidationError указывает на невалидные входные данные для оптимизации.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("invalid input: %s", e.Message)
}

