package main

import (
	"os"
	"shortleak/config"
	"shortleak/database"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	_ = os.Setenv("NODE_ENV", "test")
	_ = os.Setenv("DB_DATABASE_TEST", "shortleak-test")
	_ = os.Setenv("DB_USERNAME_TEST", "postgres")
	_ = os.Setenv("DB_PASSWORD_TEST", "12345")
	_ = os.Setenv("DB_HOST_TEST", "postgres-test")
	_ = os.Setenv("DB_DIALECT_TEST", "postgres")
	_ = os.Setenv("DB_PORT_TEST", "5432")
}

func TestRunSeed(t *testing.T) {
	calledConnect := false
	calledSeed := false

	database.ConnectDBFunc = func(cfg config.Config) {
		calledConnect = true
	}
	database.SeedFunc = func() error {
		calledSeed = true
		return nil
	}

	RunSeed()

	assert.True(t, calledConnect, "ConnectDB must be called")
	assert.True(t, calledSeed, "Seed must be called")
}

func TestMainFunc(t *testing.T) {
	calledConnect := false
	calledSeed := false

	database.ConnectDBFunc = func(cfg config.Config) {
		calledConnect = true
	}
	database.SeedFunc = func() error {
		calledSeed = true
		return nil
	}

	main()

	assert.True(t, calledConnect, "ConnectDB must be called from main")
	assert.True(t, calledSeed, "Seed must be called from main")
}
