package main

import (
	"bytes"
	"os"
	"shortleak/database"
	packages_migrate "shortleak/packages/migrate"
	"testing"

	"gorm.io/gorm"
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

// Mock migrator untuk test
type MockMigrator struct {
	MigrateCalled    bool
	RollbackCalled   bool
	SeedCalled       bool
	HasMigrationsTbl bool
	Err              error
}

func (m *MockMigrator) Migrate(db *gorm.DB) error {
	m.MigrateCalled = true
	return m.Err
}

func (m *MockMigrator) RollbackAll(db *gorm.DB) error {
	m.RollbackCalled = true
	return m.Err
}

func (m *MockMigrator) Seed() error {
	m.SeedCalled = true
	return m.Err
}

func (m *MockMigrator) HasMigrationsTable() bool {
	return m.HasMigrationsTbl
}

// helper untuk capture output stdout
func captureOutput(f func()) string {
	var buf bytes.Buffer
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = stdout
	buf.ReadFrom(r)
	return buf.String()
}

func TestRun(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		mock           *MockMigrator
		expectedOutput string
		expectMigrate  bool
		expectRollback bool
		expectSeed     bool
	}{
		{
			name:           "NoArgs",
			args:           []string{},
			mock:           &MockMigrator{},
			expectedOutput: "✅ Database connected!\nUsage: go run main.go [migrate|rollback|refresh]\n",
		},
		{
			name:           "UnknownCommand",
			args:           []string{"unknown"},
			mock:           &MockMigrator{},
			expectedOutput: "✅ Database connected!\nUnknown command: unknown\n",
		},
		{
			name:           "Migrate",
			args:           []string{"migrate"},
			mock:           &MockMigrator{},
			expectedOutput: "✅ Database connected!\n✅ Migration success!\n",
			expectMigrate:  true,
		},
		{
			name:           "Rollback",
			args:           []string{"rollback"},
			mock:           &MockMigrator{},
			expectedOutput: "✅ Database connected!\n✅ Rollback success!\n",
			expectRollback: true,
		},
		{
			name:           "Refresh",
			args:           []string{"refresh"},
			mock:           &MockMigrator{HasMigrationsTbl: true},
			expectedOutput: "✅ Database connected!\n✅ Migration success!\n✅ Seeding success!\n",
			expectMigrate:  true,
			expectRollback: true,
			expectSeed:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				packages_migrate.Run(tt.args, tt.mock)
			})

			if output != tt.expectedOutput {
				t.Errorf("expected %q, got: %q", tt.expectedOutput, output)
			}

			// check flags dari mock
			if tt.expectMigrate && !tt.mock.MigrateCalled {
				t.Errorf("expected Migrate to be called")
			}
			if tt.expectRollback && !tt.mock.RollbackCalled {
				t.Errorf("expected RollbackAll to be called")
			}
			if tt.expectSeed && !tt.mock.SeedCalled {
				t.Errorf("expected Seed to be called")
			}
		})
	}
}

func TestMainFuncWithRefresh(t *testing.T) {
	mock := &MockMigrator{HasMigrationsTbl: true}

	output := captureOutput(func() {
		RunMigrate([]string{"refresh"}, mock)
	})

	expected := "✅ Database connected!\n✅ Migration success!\n✅ Seeding success!\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}

	if !mock.MigrateCalled {
		t.Errorf("expected Migrate to be called")
	}
	if !mock.RollbackCalled {
		t.Errorf("expected RollbackAll to be called")
	}
	if !mock.SeedCalled {
		t.Errorf("expected Seed to be called")
	}
}

func TestMainFuncCoversMain(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"main", "refresh"}

	mock := &MockMigrator{HasMigrationsTbl: true}

	// override newMigrator agar main() tidak pakai DefaultMigrator
	oldNewMigrator := newMigrator
	newMigrator = func() database.Migrator {
		return mock
	}
	defer func() { newMigrator = oldNewMigrator }()

	output := captureOutput(func() {
		main() // ini sekarang pakai MockMigrator, bukan DB asli
	})

	expected := "✅ Database connected!\n✅ Migration success!\n✅ Seeding success!\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}

	if !mock.MigrateCalled || !mock.RollbackCalled || !mock.SeedCalled {
		t.Errorf("expected all migrator functions to be called from main()")
	}
}

func TestNewMigratorDefault(t *testing.T) {
	m := newMigrator()
	if m == nil {
		t.Errorf("expected DefaultMigrator instance, got nil")
	}
}
