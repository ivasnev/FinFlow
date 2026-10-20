package auth

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// TestPublicKeyFailureReusesConnection проверяет, что повторные 503 не оставляют соединения занятыми.
func TestPublicKeyFailureReusesConnection(t *testing.T) {
	var connections atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, "temporarily unavailable")
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connections.Add(1)
		}
	}
	server.Start()
	t.Cleanup(server.Close)
	transport := &http.Transport{}
	t.Cleanup(transport.CloseIdleConnections)
	client := NewClient(server.URL, time.Minute)
	client.httpClient.Transport = transport
	for i := 0; i < 8; i++ {
		if _, err := client.GetPublicKey(); err == nil {
			t.Fatal("503 must not produce a public key")
		}
	}
	if got := connections.Load(); got != 1 {
		t.Fatalf("created %d connections for eight small error responses, want 1", got)
	}
}

// trackedKeyBody отслеживает чтение и закрытие HTTP-тела, включая ошибку Reader.
type trackedKeyBody struct {
	reader io.Reader
	read   int
	closed int
}

// Read делегирует чтение и считает полученные байты.
func (b *trackedKeyBody) Read(p []byte) (int, error) {
	n, err := b.reader.Read(p)
	b.read += n
	return n, err
}

// Close фиксирует освобождение тела.
func (b *trackedKeyBody) Close() error {
	b.closed++
	return nil
}

// keyRoundTrip позволяет вернуть управляемое HTTP-тело без сетевого сервера.
type keyRoundTrip func(*http.Request) (*http.Response, error)

// RoundTrip выполняет подставленный транспорт теста.
func (f keyRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// failedKeyReader воспроизводит ошибку чтения уже полученного HTTP-ответа.
type failedKeyReader struct{}

// Read возвращает ошибку источника до декодирования ключа.
func (failedKeyReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

// TestPublicKeyAlwaysClosesBody проверяет успех, некорректный ключ, ошибку чтения и ограниченное чтение 503.
func TestPublicKeyAlwaysClosesBody(t *testing.T) {
	for _, tc := range []struct {
		name    string
		status  int
		reader  io.Reader
		wantErr bool
	}{
		{"success", 200, bytes.NewBufferString(base64.StdEncoding.EncodeToString(make([]byte, 32))), false},
		{"malformed key", 200, bytes.NewBufferString("invalid!"), true},
		{"read failure", 200, failedKeyReader{}, true},
		{"large error", 503, bytes.NewReader(make([]byte, 1<<20)), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &trackedKeyBody{reader: tc.reader}
			client := NewClient("http://auth.test", time.Minute)
			client.httpClient.Transport = keyRoundTrip(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: body, Header: make(http.Header)}, nil
			})
			_, err := client.GetPublicKey()
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, want error %v", err, tc.wantErr)
			}
			if body.closed != 1 {
				t.Fatalf("body closed %d times, want 1", body.closed)
			}
			if tc.status != 200 && body.read > 4096 {
				t.Fatalf("read %d error bytes, exceeds limit", body.read)
			}
		})
	}
}
