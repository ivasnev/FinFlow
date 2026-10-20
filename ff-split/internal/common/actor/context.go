package actor

import "context"

type key struct{}

// WithID передаёт подтверждённый авторизацией внешний ID в слой хранения.
func WithID(ctx context.Context, id int64) context.Context { return context.WithValue(ctx, key{}, id) }

// ID возвращает автора операции; фоновые операции не выдают себя за пользователя.
func ID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(key{}).(int64)
	return id, ok && id > 0
}
