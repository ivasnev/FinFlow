package mockserver

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// handleIdentity реализует состояние профилей для интеграционных тестов auth; вызывается под ms.mu.
func (ms *MockServer) handleIdentity(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodGet && r.URL.Path == "/api/v1/internal/identity" {
		id, _ := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
		login := r.URL.Query().Get("login")
		kind := r.URL.Query().Get("kind")
		for key, profile := range ms.profiles {
			match := false
			if login != "" {
				switch kind {
				case "email":
					match = profile["email"] == login
				case "nickname":
					match = profile["nickname"] == login
				default:
					if strings.Contains(login, "@") {
						match = profile["email"] == login
					} else {
						match = profile["nickname"] == login
					}
				}
			}
			if key == id || match {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(profile)
				return true
			}
		}
		w.WriteHeader(http.StatusNotFound)
		return true
	}
	if r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/v1/internal/identity/") {
		id, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/internal/identity/"), 10, 64)
		profile, ok := ms.profiles[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return true
		}
		var update map[string]interface{}
		if json.NewDecoder(r.Body).Decode(&update) != nil {
			w.WriteHeader(http.StatusBadRequest)
			return true
		}
		for otherID, other := range ms.profiles {
			if otherID == id {
				continue
			}
			for key, value := range update {
				if (key == "email" || key == "nickname") && value == other[key] {
					w.WriteHeader(http.StatusConflict)
					return true
				}
			}
		}
		for key, value := range update {
			if key == "email" || key == "nickname" {
				profile[key] = value
			}
		}
		profile["updated_at"] = time.Now().Unix()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(profile)
		return true
	}
	return false
}
