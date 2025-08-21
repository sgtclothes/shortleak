package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigWithEnvVars(t *testing.T) {
	// setup env
	os.Setenv("NODE_ENV", "test")
	os.Setenv("DB_DATABASE_TEST", "shortleak-test")
	os.Setenv("DB_USERNAME_TEST", "postgres")
	os.Setenv("DB_PASSWORD_TEST", "12345")
	os.Setenv("DB_HOST_TEST", "postgres-test")
	os.Setenv("DB_DIALECT_TEST", "postgres")
	os.Setenv("DB_PORT_TEST", "5433")

	defer func() {
		os.Unsetenv("NODE_ENV")
		os.Unsetenv("DB_DATABASE_TEST")
		os.Unsetenv("DB_USERNAME_TEST")
		os.Unsetenv("DB_PASSWORD_TEST")
		os.Unsetenv("DB_HOST_TEST")
		os.Unsetenv("DB_DIALECT_TEST")
		os.Unsetenv("DB_PORT_TEST")
	}()

	cfg := LoadConfig()

	assert.Equal(t, "test", cfg.Env)
	assert.Equal(t, "shortleak-test", cfg.Database)
	assert.Equal(t, "postgres", cfg.User)
	assert.Equal(t, "12345", cfg.Password)
	assert.Equal(t, "postgres-test", cfg.Host)
	assert.Equal(t, "postgres", cfg.Dialect)
	assert.Equal(t, "5432", cfg.Port)
}

func TestLoadConfigDefaultValues(t *testing.T) {
	// kosongin env
	os.Clearenv()

	// kasih minimal DB_DATABASE_DEVELOPMENT supaya tidak fatal
	os.Setenv("DB_DATABASE_DEVELOPMENT", "shortleak-dev")

	defer os.Clearenv()

	cfg := LoadConfig()

	assert.Equal(t, "development", cfg.Env)
	assert.Equal(t, "shortleak-dev", cfg.Database)
	assert.Equal(t, "postgres", cfg.Dialect) // fallback default
	assert.Equal(t, "5432", cfg.Port)        // fallback default
}

func TestLoadConfigMissingDatabaseShouldFatal(t *testing.T) {
	// kosongin env
	os.Clearenv()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected LoadConfig to fatal when DB_DATABASE missing")
		}
	}()

	// override log.Fatalf untuk panic (biar bisa ditest)
	oldFatal := LogFatalf
	LogFatalf = func(format string, v ...interface{}) {
		panic("fatal called")
	}
	defer func() { LogFatalf = oldFatal }()

	LoadConfig()
}

func TestToUpperEmptyString(t *testing.T) {
	result := toUpper("")
	assert.Equal(t, "", result, "expected empty string if input empty")
}

func TestToUpperNormal(t *testing.T) {
	result := toUpper("test")
	assert.Equal(t, "Test", result)
}
