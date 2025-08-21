package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"shortleak/database"
	"shortleak/dto"
	"shortleak/models"
	"shortleak/utils"
	"testing"

	"github.com/dyatlov/go-opengraph/opengraph"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type MockValidator struct {
	validateURLDirectErr error
	validateURLFormatErr error
	validateStructErr    error
}

func (m MockValidator) ValidateUrlDirect(url string) error {
	return m.validateURLDirectErr
}
func (m MockValidator) ValidateUrlFormatDirect(url string) error {
	return m.validateURLFormatErr
}
func (m MockValidator) ValidateStruct(req interface{}) error {
	return m.validateStructErr
}
func (m MockValidator) GenerateRandomString(n int) string {
	return "abcde"
}

func setupTestLinkDB(t *testing.T) {
	dsn := "host=postgres-test user=postgres password=12345 dbname=shortleak-test port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	if os.Getenv("DB_DATABASE_TEST") != "" {
		dsn = os.Getenv("DB_DATABASE_TEST")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect test DB: %v", err)
	}

	// bersihkan tabel agar fresh
	err = db.Migrator().DropTable(&models.User{}, &models.Log{}, &models.Link{})
	if err != nil {
		t.Fatalf("failed to drop tables: %v", err)
	}

	// migrasi ulang tabel
	err = db.AutoMigrate(&models.User{}, &models.Log{}, &models.Link{})
	if err != nil {
		t.Fatalf("failed to migrate test DB: %v", err)
	}

	database.DB = db
}

func createTestUser(t *testing.T) models.User {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("Secret123!"), bcrypt.DefaultCost)
	user := models.User{
		FullName: "John Doe",
		Email:    "john@example.com",
		Password: string(hashed),
		Active:   true,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
	return user
}

func TestCreateLinkSuccess(t *testing.T) {
	setupTestLinkDB(t)
	user := createTestUser(t)

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// inject user ke context pakai middleware
	r.POST("/links", func(c *gin.Context) {
		c.Set("user", user)
		CreateLink(c)
	})

	body := map[string]string{
		"url": "https://google.com",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "shortToken")
}

func TestCreateLinkMissingURL(t *testing.T) {
	setupTestLinkDB(t)
	user := createTestUser(t)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/links", func(c *gin.Context) {
		c.Set("user", user)
		CreateLink(c)
	})

	body := map[string]string{
		"url": "",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing URL")
}

func TestGetLinksByUserAuthSuccess(t *testing.T) {
	setupTestLinkDB(t)
	user := createTestUser(t)

	// buat link untuk user
	link := models.Link{URL: "https://example.com", UserID: user.ID, ShortToken: "abc12"}
	database.DB.Create(&link)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/links", func(c *gin.Context) {
		c.Set("user", user)
		GetLinksByUserAuth(c)
	})

	req, _ := http.NewRequest("GET", "/links", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "example.com")
}

func TestGetLinksByUserAuthUnauthorized(t *testing.T) {
	setupTestLinkDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/links", GetLinksByUserAuth)

	req, _ := http.NewRequest("GET", "/links", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Unauthorized")
}

func TestGetLinksByUserAuthServiceError(t *testing.T) {
	// backup
	orig := getLinksByUserID
	// override jadi error
	getLinksByUserID = func(_ uuid.UUID) ([]models.Link, error) {
		return nil, errors.New("mock error")
	}
	defer func() { getLinksByUserID = orig }()

	user := models.User{ID: uuid.New()}

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/links", func(c *gin.Context) {
		c.Set("user", user)
		GetLinksByUserAuth(c)
	})

	req, _ := http.NewRequest("GET", "/links", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid userID format")
}

func TestGetLinkByShortToken(t *testing.T) {
	setupTestLinkDB(t)

	user := createTestUser(t)

	// buat link untuk user
	link := models.Link{URL: "https://example.com", UserID: user.ID, ShortToken: "abc12"}
	database.DB.Create(&link)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/links/:shortToken", GetLinkByShortToken)

	req, _ := http.NewRequest("GET", "/links/abc12", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "https://example.com")
}

func TestGetLinkByShortTokenNotFound(t *testing.T) {
	setupTestLinkDB(t)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/links/:shortToken", GetLinkByShortToken)

	req, _ := http.NewRequest("GET", "/links/xxxx", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Link not found")
}

func TestDeleteLinkSuccess(t *testing.T) {
	setupTestLinkDB(t)
	user := createTestUser(t)

	link := models.Link{URL: "https://example.com", UserID: user.ID, ShortToken: "del12"}
	database.DB.Create(&link)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.DELETE("/links/:shortToken", DeleteLink)

	req, _ := http.NewRequest("DELETE", "/links/del12", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Link deleted successfully")
}

func TestCreateLinkInvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := bytes.NewBufferString(`{invalid-json}`)
	c.Request, _ = http.NewRequest("POST", "/links", body)
	c.Request.Header.Set("Content-Type", "application/json")

	CreateLink(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateLinkInvalidFormat(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := dto.LinkRequest{URL: "invalid-url"}
	b, _ := json.Marshal(req)

	c.Request, _ = http.NewRequest("POST", "/links", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")

	// inject mock validator yang return error di ValidateUrlFormatDirect
	Validator = MockValidator{validateURLFormatErr: errors.New("invalid format")}

	CreateLink(c)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCreateLinkUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := dto.LinkRequest{URL: "http://example.com"}
	b, _ := json.Marshal(req)

	c.Request, _ = http.NewRequest("POST", "/links", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")

	Validator = MockValidator{} // semua valid
	CreateLink(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateLinkURLAlreadyExists(t *testing.T) {
	setupTestLinkDB(t)

	user := models.User{ID: uuid.New(), FullName: "Tester"}
	database.DB.Create(&user)

	existing := models.Link{URL: "http://exists.com", UserID: user.ID, ShortToken: "abcde"}
	database.DB.Create(&existing)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := dto.LinkRequest{URL: "http://exists.com"}
	b, _ := json.Marshal(body)
	c.Request, _ = http.NewRequest("POST", "/links", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user", user)

	CreateLink(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "abcde")
}

func TestCreateLinkGenerateUniqueShortToken(t *testing.T) {
	setupTestLinkDB(t)

	user := models.User{ID: uuid.New(), FullName: "Tester"}
	database.DB.Create(&user)

	// Simulasi token bentrok pertama kali
	existing := models.Link{URL: "http://dup.com", UserID: user.ID, ShortToken: "xxxxx"}
	database.DB.Create(&existing)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := dto.LinkRequest{URL: "http://new.com"}
	b, _ := json.Marshal(body)
	c.Request, _ = http.NewRequest("POST", "/links", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user", user)

	CreateLink(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	// Pastikan token berbeda dari existing
	assert.NotContains(t, w.Body.String(), "xxxxx")
}

func TestCreateLinkServiceError(t *testing.T) {
	setupTestLinkDB(t)

	user := models.User{ID: uuid.New(), FullName: "Tester"}
	database.DB.Create(&user)

	// Override services.CreateLink sementara
	orig := createLink
	createLink = func(link *models.Link) error {
		return errors.New("service error")
	}
	defer func() { createLink = orig }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := dto.LinkRequest{URL: "http://error.com"}
	b, _ := json.Marshal(body)
	c.Request, _ = http.NewRequest("POST", "/links", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user", user)

	CreateLink(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "service error")
}

func TestCreateLinkShortTokenCollision(t *testing.T) {
	setupTestLinkDB(t)

	user := models.User{ID: uuid.New(), FullName: "Tester"}
	database.DB.Create(&user)

	// Buat link existing dengan shortToken yang kemungkinan di-generate ulang
	existing := models.Link{
		URL:        "http://existing.com",
		UserID:     user.ID,
		ShortToken: "ABCDE", // nanti kita paksa GenerateRandomString return ini dulu
	}
	database.DB.Create(&existing)

	// Override GenerateRandomString biar bentrok dulu
	origGen := utils.GenerateRandomString
	defer func() { utils.GenerateRandomString = origGen }()

	calls := 0
	utils.GenerateRandomString = func(n int) string {
		calls++
		if calls == 1 {
			return "ABCDE" // bentrok
		}
		return "NEW12" // yang valid
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := dto.LinkRequest{URL: "http://new.com"}
	b, _ := json.Marshal(body)
	c.Request, _ = http.NewRequest("POST", "/links", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user", user)

	CreateLink(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "NEW12")
}

func TestCreateLinkLogSaveError(t *testing.T) {
	dsn := "host=postgres-test user=postgres password=12345 dbname=shortleak-test port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	if os.Getenv("DB_DATABASE_TEST") != "" {
		dsn = os.Getenv("DB_DATABASE_TEST")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect test DB: %v", err)
	}

	// bersihkan tabel agar fresh
	err = db.Migrator().DropTable(&models.User{}, &models.Log{}, &models.Link{})
	if err != nil {
		t.Fatalf("failed to drop tables: %v", err)
	}

	// migrasi ulang tabel
	err = db.AutoMigrate(&models.User{}, &models.Link{})
	if err != nil {
		t.Fatalf("failed to migrate test DB: %v", err)
	}

	database.DB = db
	user := models.User{ID: uuid.New(), FullName: "Tester"}
	database.DB.Create(&user)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := dto.LinkRequest{URL: "http://logfail.com"}
	b, _ := json.Marshal(body)
	c.Request, _ = http.NewRequest("POST", "/links", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user", user)

	CreateLink(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "does not exist")
}

func TestRedirectLinkNotFound(t *testing.T) {
	setupTestLinkDB(t)

	// Override GetLinkByShortToken
	orig := getLinkByShortToken
	getLinkByShortToken = func(token string) (*models.Link, error) {
		return nil, errors.New("not found")
	}
	defer func() { getLinkByShortToken = orig }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "shortToken", Value: "abcde"}}

	RedirectLink(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Link not found")
}

func TestRedirectLinkOGError(t *testing.T) {
	setupTestLinkDB(t)

	// mock link service return success
	orig := getLinkByShortToken
	getLinkByShortToken = func(token string) (*models.Link, error) {
		return &models.Link{
			URL: "http://example.com",
		}, nil
	}
	defer func() { getLinkByShortToken = orig }()

	// override OG parser -> return error
	origOG := getOpenGraphData
	getOpenGraphData = func(url string) (opengraph.OpenGraph, error) {
		return opengraph.OpenGraph{}, errors.New("og error")
	}
	defer func() { getOpenGraphData = origOG }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "shortToken", Value: "abcde"}}

	RedirectLink(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "og error")
}

func TestRedirectLinkInvalidClientID(t *testing.T) {
	setupTestLinkDB(t)

	// mock link + OG data ok
	getLinkByShortToken = func(token string) (*models.Link, error) {
		return &models.Link{URL: "http://example.com"}, nil
	}
	getOpenGraphData = func(url string) (opengraph.OpenGraph, error) {
		return opengraph.OpenGraph{
			Title:       "Example Title",
			Description: "Example Description",
			URL:         url,
		}, nil
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "shortToken", Value: "abcde"}}
	c.Request, _ = http.NewRequest("GET", "/links/abcde", nil)
	// invalid uuid string
	c.Request.AddCookie(&http.Cookie{Name: "client_id", Value: "not-a-uuid"})

	RedirectLink(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid client ID")
}

func setupTestLinkDBNoLogs(t *testing.T) {
	dsn := "host=postgres-test user=postgres password=12345 dbname=shortleak-test port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect test DB: %v", err)
	}

	// drop semua tabel
	_ = db.Migrator().DropTable(&models.User{}, &models.Link{}, &models.Log{})

	// migrasi hanya User & Link, jangan Log
	if err := db.AutoMigrate(&models.User{}, &models.Link{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	database.DB = db
}

func TestRedirectLinkLogSaveError(t *testing.T) {
	// pakai setup tanpa logs
	setupTestLinkDBNoLogs(t)

	getLinkByShortToken = func(token string) (*models.Link, error) {
		return &models.Link{URL: "http://example.com"}, nil
	}
	getOpenGraphData = func(url string) (opengraph.OpenGraph, error) {
		return opengraph.OpenGraph{
			Title:       "Example Title",
			Description: "Example Description",
			URL:         url,
		}, nil
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "shortToken", Value: "abcde"}}
	c.Request, _ = http.NewRequest("GET", "/links/abcde", nil)
	c.Request.AddCookie(&http.Cookie{Name: "client_id", Value: uuid.New().String()})

	RedirectLink(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "does not exist")
}

func TestRedirectLinkSuccess(t *testing.T) {
	setupTestLinkDB(t)

	getLinkByShortToken = func(token string) (*models.Link, error) {
		return &models.Link{URL: "http://example.com"}, nil
	}
	getOpenGraphData = func(url string) (opengraph.OpenGraph, error) {
		return opengraph.OpenGraph{
			Title:       "Example Title",
			Description: "Example Description",
			URL:         url,
		}, nil
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "shortToken", Value: "abcde"}}
	c.Request, _ = http.NewRequest("GET", "/links/abcde", nil)
	c.Request.AddCookie(&http.Cookie{Name: "client_id", Value: uuid.New().String()})

	RedirectLink(c)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "http://example.com", w.Header().Get("Location"))
}

func TestGetLinkStatsNotFound(t *testing.T) {
	setupTestLinkDB(t)

	// override service untuk balikin error
	orig := getLinkByShortToken
	getLinkByShortToken = func(token string) (*models.Link, error) {
		return nil, errors.New("not found")
	}
	defer func() { getLinkByShortToken = orig }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "shortToken", Value: "abcde"}}

	GetLinkStats(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Link not found")
}

func TestGetLinkStatsDBErrorTotalVisits(t *testing.T) {
	setupTestLinkDB(t)

	// link ditemukan
	orig := getLinkByShortToken
	getLinkByShortToken = func(token string) (*models.Link, error) {
		return &models.Link{ShortToken: token, URL: "http://example.com"}, nil
	}
	defer func() { getLinkByShortToken = orig }()

	// pakai DB rusak → drop tabel logs biar query gagal
	_ = database.DB.Migrator().DropTable(&models.Log{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "shortToken", Value: "abcde"}}

	GetLinkStats(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "does not exist")
}

func TestCountUniqueVisitorsSuccess(t *testing.T) {
	// pakai DB test
	setupTestLinkDB(t)
	database.DB.Exec("DELETE FROM logs")

	// Insert 3 logs -> 2 user berbeda (jadi hasilnya harus 2)
	userID1 := uuid.New()
	userID2 := uuid.New()
	database.DB.Create(&models.Log{
		Action: "visit-link",
		UserID: userID1,
		Data:   datatypes.JSON([]byte(`{"shortToken":"abcde"}`)),
	})
	database.DB.Create(&models.Log{
		Action: "visit-link",
		UserID: userID2,
		Data:   datatypes.JSON([]byte(`{"shortToken":"abcde"}`)),
	})
	database.DB.Create(&models.Log{
		Action: "visit-link",
		UserID: userID1,
		Data:   datatypes.JSON([]byte(`{"shortToken":"abcde"}`)),
	})

	uniqueVisitors, err := countUniqueVisitors("abcde")

	assert.NoError(t, err)
	assert.Equal(t, int64(2), uniqueVisitors)
}

func TestCountUniqueVisitorsError(t *testing.T) {
	// Simpan asli
	origCount := countUniqueVisitors
	// Override untuk simulate error
	countUniqueVisitors = func(shortToken string) (int64, error) {
		return 0, errors.New("mock count unique visitors error")
	}
	defer func() { countUniqueVisitors = origCount }()

	uniqueVisitors, err := countUniqueVisitors("abcde")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock count unique visitors error")
	assert.Equal(t, int64(0), uniqueVisitors)
}

func TestGetLinkStatsDBErrorUniqueVisitors(t *testing.T) {
	setupTestLinkDB(t)

	// Mock getLinkByShortToken biar return link valid
	origLink := getLinkByShortToken
	getLinkByShortToken = func(token string) (*models.Link, error) {
		return &models.Link{
			ShortToken: "abcde",
			URL:        "http://example.com",
		}, nil
	}
	defer func() { getLinkByShortToken = origLink }()

	// Mock countUniqueVisitors biar return error
	origCount := countUniqueVisitors
	countUniqueVisitors = func(shortToken string) (int64, error) {
		return 0, errors.New("mock unique visitors error")
	}
	defer func() { countUniqueVisitors = origCount }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "shortToken", Value: "abcde"}}

	GetLinkStats(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "mock unique visitors error")
}

func TestGetLinkStatsSuccess(t *testing.T) {
	setupTestLinkDB(t)

	// link ditemukan
	orig := getLinkByShortToken
	getLinkByShortToken = func(token string) (*models.Link, error) {
		return &models.Link{ShortToken: token, URL: "http://example.com"}, nil
	}
	defer func() { getLinkByShortToken = orig }()

	// buat dummy logs
	user1 := uuid.New()
	user2 := uuid.New()
	logs := []models.Log{
		{UserID: user1, Action: "visit-link", Data: datatypes.JSON([]byte(`{"shortToken":"abcde"}`))},
		{UserID: user1, Action: "visit-link", Data: datatypes.JSON([]byte(`{"shortToken":"abcde"}`))},
		{UserID: user2, Action: "visit-link", Data: datatypes.JSON([]byte(`{"shortToken":"abcde"}`))},
	}
	for _, l := range logs {
		_ = database.DB.Create(&l).Error
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "shortToken", Value: "abcde"}}

	GetLinkStats(c)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "http://example.com")
	assert.Contains(t, body, `"totalVisits":3`)
	assert.Contains(t, body, `"uniqueVisitors":2`)
}

func TestDeleteLinkError(t *testing.T) {
	// Simpan original service
	origDeleteLink := deleteLink
	// Mock supaya error
	deleteLink = func(shortToken string) error {
		return errors.New("delete failed")
	}
	defer func() { deleteLink = origDeleteLink }()

	// Setup router
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodDelete, "/links/abcde", nil)
	c.Request = req
	c.Params = []gin.Param{{Key: "shortToken", Value: "abcde"}}

	// Call handler
	DeleteLink(c)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "delete failed")
}
