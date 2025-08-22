package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"shortleak/database"
	"shortleak/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestAuthDB(t *testing.T) {
	dsn := "host=localhost user=postgres password=12345 dbname=shortleak-test port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	if os.Getenv("DB_DATABASE_TEST") != "" {
		dsn = os.Getenv("DB_DATABASE_TEST")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect test DB: %v", err)
	}

	// bersihkan tabel agar fresh
	err = db.Migrator().DropTable(&models.User{}, &models.Log{})
	if err != nil {
		t.Fatalf("failed to drop tables: %v", err)
	}

	// migrasi ulang tabel
	err = db.AutoMigrate(&models.User{}, &models.Log{})
	if err != nil {
		t.Fatalf("failed to migrate test DB: %v", err)
	}

	database.DB = db
}

func TestRegisterInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Router sementara
	r := gin.Default()
	r.POST("/register", Register)

	// Kirim request dengan body yang bukan JSON valid
	reqBody := `{"fullname": "Test User", "email": "test@example.com", "password": "12345"`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

func TestRegisterValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.Default()
	r.POST("/register", Register)

	// Kirim JSON valid format tapi gagal validasi (misal email kosong)
	reqBody := `{"fullname": "", "email": "invalid-email", "password": ""}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "validation_errors")
}

func TestRegisterUserAlreadyExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestAuthDB(t)

	// Bersihkan tabel users
	database.DB.Exec("DELETE FROM users")

	// Insert user dummy
	existing := models.User{
		FullName: "Existing User",
		Email:    "existing@example.com",
		Password: "hashedpassword",
	}
	database.DB.Create(&existing)

	// Setup router
	r := gin.Default()
	r.POST("/register", Register)

	// Request dengan email sama
	reqBody := `{"fullname": "Another User", "email": "existing@example.com", "password": "Secret123!"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "User already exists")
}

func TestRegisterBcryptError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	setupTestAuthDB(t)

	// Mock bcrypt supaya return error
	generatePasswordHash = func(password []byte, cost int) ([]byte, error) {
		return nil, errors.New("bcrypt error")
	}
	defer func() {
		generatePasswordHash = bcrypt.GenerateFromPassword // restore asli
	}()

	// Setup router
	r := gin.Default()
	r.POST("/register", Register)

	// Request valid supaya masuk ke bcrypt
	reqBody := `{"fullname":"John Doe","email":"john@example.com","password":"Secret123!"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Expect gagal
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to register user")
	assert.Contains(t, w.Body.String(), "bcrypt error")
}

func TestRegisterDBInsertUserFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	gdb, _ := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})

	database.DB = gdb

	r := gin.Default()
	r.POST("/register", Register)

	// siapkan expectation: insert ke "users" gagal
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	body := `{"fullname": "John", "email": "john@example.com", "password": "Secret123!"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to register user")
}

func TestRegisterDBInsertLogFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	gdb, _ := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})

	database.DB = gdb

	r := gin.Default()
	r.POST("/register", Register)

	// Transaction begin
	mock.ExpectBegin()

	// Insert user sukses (pakai Query RETURNING id)
	mock.ExpectQuery(`INSERT INTO "users"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "John", "johnlogfail@example.com", sqlmock.AnyArg(), true, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("f466fb51-1aff-46c0-bfaf-d3e3931582c9"))

	// Insert log gagal
	mock.ExpectQuery(`INSERT INTO "logs"`).
		WillReturnError(errors.New("insert log failed"))

	// Karena gagal → rollback
	mock.ExpectRollback()

	body := `{"fullname": "John", "email": "johnlogfail@example.com", "password": "Secret123!"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to register user")
	assert.Contains(t, w.Body.String(), "insert log failed")

	// pastikan semua expectation kepanggil
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestRegisterSuccess(t *testing.T) {
	setupTestAuthDB(t)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/register", Register)

	body := map[string]string{
		"fullname": "John Doe",
		"email":    "john@example.com",
		"password": "Secret123!",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "User registered successfully")
}

func TestLoginSuccess(t *testing.T) {
	setupTestAuthDB(t)
	os.Setenv("PLATFORM", "token")

	// buat user dengan password bcrypt
	hashed, _ := bcrypt.GenerateFromPassword([]byte("Secret123!"), bcrypt.DefaultCost)
	user := models.User{
		FullName: "John Doe",
		Email:    "john@example.com",
		Password: string(hashed),
		Active:   true,
	}
	database.DB.Create(&user)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/login", Login)

	body := map[string]string{
		"email":    "john@example.com",
		"password": "Secret123!",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Login successful")
	assert.Contains(t, w.Body.String(), "token")
}

func TestLoginInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Router sementara
	r := gin.Default()
	r.POST("/login", Login)

	// Kirim request dengan body yang bukan JSON valid
	reqBody := `{"email": "test@example.com", "password": "12345"`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

func TestLoginValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.Default()
	r.POST("/login", Login)

	// Kirim JSON valid format tapi gagal validasi (misal email kosong)
	reqBody := `{"email": "invalid-email", "password": ""}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "validation_errors")
}

func TestLoginDBInsertUserFails(t *testing.T) {
	// konek ke Postgres test
	dsn := "host=localhost user=postgres password=12345 dbname=shortleak-test port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed connect to test db: %v", err)
	}
	database.DB = gdb

	// pastikan migration jalan
	if err := gdb.AutoMigrate(&models.User{}, &models.Log{}); err != nil {
		t.Fatalf("failed migrate: %v", err)
	}

	// bersihin dulu
	gdb.Exec("TRUNCATE users RESTART IDENTITY CASCADE")
	gdb.Exec("TRUNCATE logs RESTART IDENTITY CASCADE")

	// insert dummy user
	hashed, _ := bcrypt.GenerateFromPassword([]byte("Secret123!"), bcrypt.DefaultCost)
	user := models.User{
		FullName: "John Doe",
		Email:    "john@example.com",
		Password: string(hashed),
		Active:   true,
	}
	if err := gdb.Create(&user).Error; err != nil {
		t.Fatalf("failed insert dummy user: %v", err)
	}

	// override DB Create for logs to always fail using GORM callback
	origDB := database.DB
	database.DB.Callback().Create().Before("gorm:create").Register("force_log_create_error", func(db *gorm.DB) {
		if db.Statement.Table == "logs" {
			db.AddError(errors.New("insert failed"))
		}
	})

	// setup gin
	r := gin.Default()
	r.POST("/login", Login)

	body := `{"email": "john@example.com", "password": "Secret123!"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to log login attempt")

	// cleanup user biar DB test bersih
	gdb.Exec("TRUNCATE users RESTART IDENTITY CASCADE")
	gdb.Exec("TRUNCATE logs RESTART IDENTITY CASCADE")

	// balikin DB normal
	database.DB = origDB
}

func TestLoginTokenSigningFails(t *testing.T) {
	setupTestAuthDB(t)
	origSigner := signToken
	signToken = func(_ *jwt.Token, _ interface{}) (string, error) {
		return "", errors.New("signing failed")
	}
	defer func() { signToken = origSigner }()

	// gunakan database asli (sudah dikonfigurasi di init test)
	// bersihkan dulu user dengan email test biar tidak bentrok
	database.DB.Exec(`DELETE FROM users WHERE email = ?`, "john@example.com")

	// buat user dummy
	hashed, _ := bcrypt.GenerateFromPassword([]byte("Secret123!"), bcrypt.DefaultCost)
	user := models.User{
		ID:       uuid.New(),
		FullName: "John Doe",
		Email:    "john@example.com",
		Password: string(hashed),
		Active:   true,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	// cleanup user setelah test
	defer database.DB.Exec(`DELETE FROM users WHERE email = ?`, "john@example.com")

	// setup gin router
	r := gin.Default()
	r.POST("/login", Login)

	body := `{"email":"john@example.com","password":"Secret123!"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to create token")
}

func TestLoginInvalidPassword(t *testing.T) {
	setupTestAuthDB(t)
	os.Setenv("PLATFORM", "token")

	// buat user
	hashed, _ := bcrypt.GenerateFromPassword([]byte("Secret123!"), bcrypt.DefaultCost)
	user := models.User{
		FullName: "Jane Doe",
		Email:    "jane@example.com",
		Password: string(hashed),
		Active:   true,
	}
	database.DB.Create(&user)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/login", Login)

	body := map[string]string{
		"email":    "jane@example.com",
		"password": "WrongPass!",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid email or password")
}

func TestLoginUserNotFound(t *testing.T) {
	setupTestAuthDB(t)
	os.Setenv("PLATFORM", "token")

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/login", Login)

	body := map[string]string{
		"email":    "nouser@example.com",
		"password": "Secret123!",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid email or password")
}

func TestLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/logout", Logout)

	req, _ := http.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Logged out")

	// pastikan cookie token terhapus
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "token" && c.Value == "" && c.MaxAge < 0 {
			found = true
		}
	}
	assert.True(t, found, "Logout should clear token cookie")
}
