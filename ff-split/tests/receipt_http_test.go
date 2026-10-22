package tests

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	receiptrepo "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/receipt"
	receiptservice "github.com/ivasnev/FinFlow/ff-split/internal/service/receipt"
)

// receiptHTTPStore хранит только синтетические фотографии и считает обращения к хранилищу.
type receiptHTTPStore struct {
	mu          sync.Mutex
	objects     map[string][]byte
	puts, reads int
}

// Put сохраняет ограниченные синтетические байты.
func (s *receiptHTTPStore) Put(_ context.Context, k string, b []byte, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.puts++
	s.objects[k] = append([]byte{}, b...)
	return nil
}

// Read читает копию синтетического изображения.
func (s *receiptHTTPStore) Read(_ context.Context, k string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reads++
	b, ok := s.objects[k]
	if !ok {
		return nil, fmt.Errorf("missing synthetic object")
	}
	return append([]byte{}, b...), nil
}

// Delete удаляет синтетический объект.
func (s *receiptHTTPStore) Delete(_ context.Context, k string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects, k)
	return nil
}

// countReceiptReader выявляет чтение тела до проверки авторизации или занятого слота.
type countReceiptReader struct{ reads int }

// Read выявляет нежелательное чтение тела.
func (r *countReceiptReader) Read([]byte) (int, error) { r.reads++; return 0, io.EOF }

// TestReceiptHTTPBoundary проверяет настоящий маршрутизатор, JWT, лимиты и границы мероприятия.
// Использует штатные fixture settlementFixture и receiptMigration из пакета tests.
func TestReceiptHTTPBoundary(t *testing.T) {
	database, c := settlementFixture(t)
	receiptMigration(t, database)
	if err := database.DB.Exec("INSERT INTO user_event(user_id,event_id) VALUES(1,2)").Error; err != nil {
		t.Fatal(err)
	}
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	keys := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, base64.StdEncoding.EncodeToString(pub)) }))
	t.Cleanup(keys.Close)
	store := &receiptHTTPStore{objects: map[string][]byte{}}
	repo := receiptrepo.NewRepository(database.DB)
	pipeline := receiptservice.NewService(repo, store, nil, c.TransactionService)
	c.AuthClient = auth.NewClient(keys.URL, time.Hour)
	c.ServerHandler = handler.NewServerHandler(c.EventService, c.UserService, c.TransactionService, c.ActivityService, c.TaskService, c.CategoryService, c.IconService, nil)
	c.RegisterRoutes()
	c.ServerHandler.SetReceiptService(pipeline)
	request := func(actor int64, purpose, method, path string, body io.Reader) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, body)
		r.Header.Set("Content-Type", "application/octet-stream")
		r.Header.Set("X-Actor-ID", "202")
		if actor > 0 {
			payload, _ := json.Marshal(map[string]any{"user_id": actor, "roles": []string{"user"}, "exp": time.Now().Add(time.Hour).Unix(), "purpose": purpose})
			token, _ := json.Marshal(auth.Token{Payload: payload, Sig: ed25519.Sign(priv, payload)})
			r.Header.Set("Authorization", "Bearer "+base64.StdEncoding.EncodeToString(token))
		}
		out := httptest.NewRecorder()
		c.Router.ServeHTTP(out, r)
		return out
	}
	check := func(actor int64, method, path string, body io.Reader, want int) *httptest.ResponseRecorder {
		t.Helper()
		out := request(actor, "access", method, path, body)
		if out.Code != want {
			t.Fatalf("%s %s actor%d got%d want%d: %s", method, path, actor, out.Code, want, out.Body)
		}
		return out
	}
	const root = "/api/v1/event/1/receipts"
	untouched := &countReceiptReader{}
	check(0, "POST", root, untouched, 401)
	if untouched.reads != 0 {
		t.Fatal("unauthenticated body read")
	}
	check(404, "POST", root, bytes.NewReader(nil), 403)
	wrongPurpose := request(101, "refresh", "GET", root, nil)
	if wrongPurpose.Code != 401 {
		t.Fatal("refresh bearer accepted")
	}
	check(101, "POST", root, strings.NewReader("not image"), 400)
	check(101, "POST", root, bytes.NewReader(make([]byte, receiptservice.MaxImageBytes+1)), 413)
	var releases []func()
	for i := 0; i < 4; i++ {
		release, ok := pipeline.ImageSlot()
		if !ok {
			t.Fatal("slot unavailable")
		}
		releases = append(releases, release)
	}
	untouched = &countReceiptReader{}
	out := check(101, "POST", root, untouched, 429)
	if untouched.reads != 0 || out.Header().Get("Retry-After") == "" {
		t.Fatal("slot rejection read body or lacked retry")
	}
	for _, release := range releases {
		release()
	}
	photo := func(n uint8) []byte {
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		img.Set(0, 0, color.RGBA{R: n, A: 255})
		var b bytes.Buffer
		if err := png.Encode(&b, img); err != nil {
			t.Fatal(err)
		}
		return b.Bytes()
	}
	out = check(101, "POST", root, bytes.NewReader(photo(1)), 202)
	if out.Header().Get("Cache-Control") != "no-store" || strings.Contains(out.Body.String(), "receipts/") || strings.Contains(out.Body.String(), "lease_token") {
		t.Fatal("receipt metadata leaked/cacheable")
	}
	var row models.Receipt
	if err = json.Unmarshal(out.Body.Bytes(), &row); err != nil {
		t.Fatal(err)
	}
	if row.CreatedBy != 101 || row.Status != "queued" {
		t.Fatal("actor spoof/status mismatch")
	}
	path := root + "/" + row.ID
	foreign := "/api/v1/event/2/receipts/" + row.ID
	for _, suffix := range []string{"", "/image"} {
		check(0, "GET", path+suffix, nil, 401)
		check(404, "GET", path+suffix, nil, 403)
		check(101, "GET", foreign+suffix, nil, 404)
	}
	if store.reads != 0 {
		t.Fatal("foreign image reached storage")
	}
	// Занятый общий пул отвергает чтение фотографии до обращения к S3.
	releases = nil
	for i := 0; i < 4; i++ {
		release, ok := pipeline.ImageSlot()
		if !ok {
			t.Fatal("image slot unavailable")
		}
		releases = append(releases, release)
	}
	out = check(202, "GET", path+"/image", nil, 429)
	if store.reads != 0 || out.Header().Get("Retry-After") == "" {
		t.Fatal("busy image GET reached storage or lacked retry")
	}
	for _, release := range releases {
		release()
	}
	out = check(202, "GET", path+"/image", nil, 200)
	if !bytes.Equal(out.Body.Bytes(), photo(1)) || out.Header().Get("Cache-Control") != "no-store" || out.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("image contract mismatch")
	}
	check(101, "GET", root+"/invalid-uuid", nil, 400)
	check(101, "GET", root+"?before=abc", nil, 400)
	valid := fmt.Sprintf(`{"revision":%d,"payer_id":1,"items":[{"name":"Purchase","amount_cents":100,"users":[2],"selected":true}],"merge":false,"acknowledge_difference":false}`, row.Revision)
	check(101, "POST", path+"/confirm", strings.NewReader(valid), 400)
	if err = database.DB.Exec("UPDATE receipts SET status='ready' WHERE id=?", row.ID).Error; err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"", "/confirm"} {
		method := "PUT"
		if suffix != "" {
			method = "POST"
		}
		check(0, method, path+suffix, strings.NewReader(valid), 401)
		check(404, method, path+suffix, strings.NewReader(valid), 403)
		check(101, method, foreign+suffix, strings.NewReader(valid), 404)
	}
	check(101, "PUT", path, strings.NewReader(valid+`{}`), 400)
	check(101, "PUT", path, strings.NewReader(strings.Replace(valid, `"merge":false`, `"merge":false,"created_by":202`, 1)), 400)
	check(101, "PUT", path, strings.NewReader(`{"items":[],"padding":"`+strings.Repeat("x", 256<<10)+`"}`), 400)
	check(101, "POST", path+"/confirm", strings.NewReader(strings.Replace(valid, `"payer_id":1`, `"payer_id":101`, 1)), 400)
	check(101, "POST", path+"/confirm", strings.NewReader(valid), 200)
	check(101, "POST", path+"/confirm", strings.NewReader(valid), 200)
	var count int64
	database.DB.Table("transactions").Count(&count)
	if count != 1 {
		t.Fatal("confirmation duplicated expense")
	}
	check(202, "DELETE", "/api/v1/event/1", nil, 200)
	check(101, "GET", path, nil, 200)
	check(101, "GET", path+"/image", nil, 200)
	check(101, "PUT", path, strings.NewReader(valid), 409)
	check(101, "POST", path+"/confirm", strings.NewReader(valid), 409)
	check(101, "POST", root, bytes.NewReader(photo(2)), 409)
	c.ServerHandler.SetReceiptService(nil)
	check(101, "GET", root, nil, 503)
	c.ServerHandler.SetReceiptService(pipeline)
	// Суточный лимит проверяется до S3, включая неудачные и архивные задания.
	if err = database.DB.Exec(`INSERT INTO receipts(id,event_id,created_by,image_key,image_sha256,content_type,status) SELECT gen_random_uuid(),2,101,'daily/'||n,lpad(n::text,64,'0'),'image/png','failed' FROM generate_series(1,50)n`).Error; err != nil {
		t.Fatal(err)
	}
	putBefore := store.puts
	check(101, "POST", "/api/v1/event/2/receipts", bytes.NewReader(photo(3)), 400)
	if store.puts != putBefore {
		t.Fatal("daily cap wrote image")
	}
	if err = database.DB.Exec("DELETE FROM receipts WHERE image_key LIKE 'daily/%'").Error; err != nil {
		t.Fatal(err)
	}
	if err = database.DB.Exec(`INSERT INTO receipts(id,event_id,created_by,image_key,image_sha256,content_type,status) SELECT gen_random_uuid(),2,999,'global/'||n,lpad(n::text,64,'0'),'image/png','queued' FROM generate_series(1,1000)n`).Error; err != nil {
		t.Fatal(err)
	}
	check(101, "POST", "/api/v1/event/2/receipts", bytes.NewReader(photo(4)), 400)
	if store.puts != putBefore {
		t.Fatal("global cap wrote image")
	}
}
