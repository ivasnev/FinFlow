package config

import "testing"

// TestAdminTokenEnvOverride проверяет настройку токена и явное отключение доступа пустой переменной.
func TestAdminTokenEnvOverride(t *testing.T) {
	for _, token := range []string{"configured-secret", ""} {
		t.Run("token="+token, func(t *testing.T) {
			t.Setenv("TVM_ADMIN_TOKEN", token)
			cfg := &Config{AdminToken: "yaml-secret"}
			if err := loadFromEnv(cfg); err != nil {
				t.Fatal(err)
			}
			if cfg.AdminToken != token {
				t.Fatal("environment did not override admin token")
			}
		})
	}
}
