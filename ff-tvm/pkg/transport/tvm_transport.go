package transport

import (
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/ivasnev/FinFlow/ff-tvm/pkg/client"
	"github.com/ivasnev/FinFlow/ff-tvm/pkg/middleware"
)

// TVMTransport - транспорт для добавления тикетов в заголовки
type TVMTransport struct {
	baseTransport http.RoundTripper
	client        *client.TVMClient
	from          int
	to            int
}

// NewTVMTransport создает новый транспорт для добавления тикетов
func NewTVMTransport(client *client.TVMClient, baseTransport http.RoundTripper, from, to int) *TVMTransport {
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

	// Формируем строку тикета в новом формате
	serviceIDStr := strconv.Itoa(t.from)
	serviceIDBase64 := base64.StdEncoding.EncodeToString([]byte(serviceIDStr))
	ticketStr := "serv:" + serviceIDBase64 + ":" + ticket

	// Добавляем заголовок
	req.Header.Set(middleware.HeaderServiceTicket, ticketStr)

	// Выполняем запрос
	return t.baseTransport.RoundTrip(req)
}
