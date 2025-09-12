package transport

import (
	"net/http"

	"github.com/ivasnev/FinFlow/ff-tvm/pkg/client"
	"github.com/ivasnev/FinFlow/ff-tvm/pkg/middleware"
)

// TVMTransport - транспорт для добавления тикетов в заголовки
type TVMTransport struct {
	baseTransport http.RoundTripper
	client        *client.TVMClient
	from          int64
	to            int64
}

// NewTVMTransport создает новый транспорт для добавления тикетов
func NewTVMTransport(client *client.TVMClient, baseTransport http.RoundTripper, from, to int64) *TVMTransport {
	return &TVMTransport{
		client:        client,
		baseTransport: baseTransport,
		from:          from,
		to:            to,
	}
}

// RoundTrip реализует интерфейс http.RoundTripper
func (t *TVMTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Генерируем тикет
	ticket, err := t.client.GenerateTicket(t.from, t.to)
	if err != nil {
		return nil, err
	}

	// Добавляем заголовки
	req.Header.Set(middleware.HeaderTicket, ticket)

	// Выполняем запрос
	return t.baseTransport.RoundTrip(req)
}
