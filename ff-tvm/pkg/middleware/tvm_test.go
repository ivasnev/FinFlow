package middleware

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/service"
)

// ticketKeys возвращает публичный ключ тестового отправителя.
type ticketKeys struct{ key string }

// GetPublicKey предоставляет ключ для проверки подписи.
func (k ticketKeys) GetPublicKey(int) (string, error) { return k.key, nil }

// TestTicketClaims проверяет назначение, срок и подпись через защищённый HTTP-маршрут.
func TestTicketClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().Unix()
	for _, tc := range []struct {
		name               string
		from, to, receiver int
		ttl                int64
		tamper, badKey     bool
		status             int
	}{
		{name: "valid", from: 7, to: 2, receiver: 2, ttl: now + 3600, status: 204},
		{name: "expired", from: 7, to: 2, receiver: 2, ttl: now - 1, status: 401},
		{name: "zeroTTL", from: 7, to: 2, receiver: 2, status: 401},
		{name: "expiresNow", from: 7, to: 2, receiver: 2, ttl: now, status: 401},
		{name: "wrongRecipient", from: 7, to: 999, receiver: 2, ttl: now + 3600, status: 401},
		{name: "wrongSender", from: 99, to: 2, receiver: 2, ttl: now + 3600, status: 401},
		{name: "unsetReceiver", from: 7, to: 0, receiver: 0, ttl: now + 3600, status: 401},
		{name: "tampered", from: 7, to: 2, receiver: 2, ttl: now + 3600, tamper: true, status: 401},
		{name: "badKey", from: 7, to: 2, receiver: 2, ttl: now + 3600, badKey: true, status: 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload := service.TicketPayload{From: tc.from, To: tc.to, TTL: tc.ttl, Metadata: "{}"}
			data, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			sig := ed25519.Sign(priv, data)
			if tc.tamper {
				sig[0] ^= 1
			}
			ticket, err := json.Marshal(service.Ticket{Payload: payload, Signature: base64.StdEncoding.EncodeToString(sig)})
			if err != nil {
				t.Fatal(err)
			}
			key := base64.StdEncoding.EncodeToString(pub)
			if tc.badKey {
				key = base64.StdEncoding.EncodeToString([]byte("bad"))
			}
			router := gin.New()
			router.GET("/protected", NewTVMMiddleware(ticketKeys{key}, tc.receiver).ValidateTicket(), func(c *gin.Context) {
				sender, ok := c.Get(ServiceIDContext)
				if !ok || sender != 7 {
					t.Error("missing authenticated sender")
				}
				c.Status(204)
			})
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set(HeaderServiceTicket, fmt.Sprintf("serv:Nw==:%s", base64.StdEncoding.EncodeToString(ticket)))
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			if res.Code != tc.status {
				t.Fatalf("status=%d want=%d body=%s", res.Code, tc.status, res.Body)
			}
		})
	}
}
