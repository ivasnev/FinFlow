package token

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
)

// TestTokenPurpose проверяет одинаковую границу назначения на сервере и в публичном клиенте.
func TestTokenPurpose(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	manager := &ED25519TokenManager{publicKey: pub, privateKey: priv}
	keys := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(base64.StdEncoding.EncodeToString(pub)))
	}))
	t.Cleanup(keys.Close)
	client := auth.NewClient(keys.URL, time.Hour)
	access, refresh, _, refreshExpiresAt, err := manager.GenerateTokenPair(1, []string{"user"}, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := manager.GenerateToken(&service.TokenPayload{UserID: 1, Exp: time.Now().Add(time.Hour).Unix()})
	if err != nil {
		t.Fatal(err)
	}
	unknown, err := manager.GenerateToken(&service.TokenPayload{Purpose: "unknown", UserID: 1, Exp: time.Now().Add(time.Hour).Unix()})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, token string
		allowed     bool
	}{
		{"access", access, true}, {"refresh", refresh, false}, {"legacy", legacy, false}, {"unknown", unknown, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, serverErr := manager.ValidateToken(tc.token)
			_, clientErr := client.ValidateToken(tc.token)
			if (serverErr == nil) != tc.allowed || (clientErr == nil) != tc.allowed {
				t.Fatalf("allowed=%t server=%v client=%v", tc.allowed, serverErr, clientErr)
			}
			r := gin.New()
			r.GET("/protected", auth.AuthMiddleware(client), func(c *gin.Context) { c.Status(204) })
			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)
			want := 401
			if tc.allowed {
				want = 204
			}
			if res.Code != want {
				t.Fatalf("status=%d want=%d", res.Code, want)
			}
		})
	}
	// Назначение и срок refresh-токена входят в подписанный payload.
	data, err := base64.StdEncoding.DecodeString(refresh)
	if err != nil {
		t.Fatal(err)
	}
	var envelope service.Token
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatal(err)
	}
	if !ed25519.Verify(pub, envelope.Payload, envelope.Sig) {
		t.Fatal("refresh signature is invalid")
	}
	var payload service.TokenPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Exp != refreshExpiresAt || payload.Purpose != "refresh" || payload.Exp <= time.Now().Add(50*time.Minute).Unix() {
		t.Fatal("invalid refresh claims")
	}
}
