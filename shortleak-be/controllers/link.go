package controllers

import (
	"encoding/json"
	"net/http"
	"shortleak/database"
	"shortleak/dto"
	"shortleak/models"
	"shortleak/services"
	"shortleak/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var getLinksByUserID = services.GetLinksByUserID
var createLink = services.CreateLink
var deleteLink = services.DeleteLink
var (
	getLinkByShortToken = services.GetLinkByShortToken
	getOpenGraphData    = utils.GetOpenGraphData
)
var countUniqueVisitors = func(shortToken string) (int64, error) {
	var uniqueVisitors int64
	err := database.DB.Model(&models.Log{}).
		Select("COUNT(DISTINCT(user_id))").
		Where("action = ?", "visit-link").
		Where("data->>'shortToken' = ?", shortToken).
		Scan(&uniqueVisitors).Error
	return uniqueVisitors, err
}
var Validator utils.Validator = utils.DefaultValidator{}
var DB *gorm.DB

func GetLinksByUserAuth(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	u := user.(models.User)
	links, err := getLinksByUserID(u.ID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid userID format"})
		return
	}
	c.JSON(http.StatusOK, links)
}

func GetLinkByShortToken(c *gin.Context) {
	shortToken := c.Param("shortToken")
	link, err := getLinkByShortToken(shortToken)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}
	c.JSON(http.StatusOK, link)
}

func CreateLink(c *gin.Context) {
	var req dto.LinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	/** Validate URL */
	if err := utils.ValidateUrlDirect(req.URL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing URL"})
		return
	}
	/** Validate URL Format */
	if err := Validator.ValidateUrlFormatDirect(req.URL); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid URL format"})
		return
	}
	/** Get user from context */
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	u := user.(models.User)
	/** Check if link already exists */
	var existingLink models.Link
	if err := database.DB.First(&existingLink, "url = ?", req.URL).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"shortToken": existingLink.ShortToken})
		return
	}
	/** Generate unique code */
	shortToken := utils.GenerateRandomString(5)
	for {
		var existingLink models.Link
		if err := database.DB.First(&existingLink, "short_token = ?", shortToken).Error; err != nil {
			break
		}
		shortToken = utils.GenerateRandomString(5)
	}
	/** Create link */
	var link = models.Link{
		URL:        req.URL,
		UserID:     u.ID,
		ShortToken: shortToken,
	}
	if err := createLink(&link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	/** Create link log */
	b, _ := json.Marshal(link)
	log := models.Log{
		UserID: u.ID,
		Action: "create-link",
		Data:   datatypes.JSON(b),
	}

	/** Save log to database */
	if err := database.DB.Create(&log).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"shortToken": link.ShortToken})
}

func RedirectLink(c *gin.Context) {
	shortToken := c.Param("shortToken")
	link, err := getLinkByShortToken(shortToken)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}
	/** Get Open Graph Data */
	ogData, err := getOpenGraphData(link.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	/** Get client id from context */
	clientID, _ := c.Cookie("client_id")
	uID, err := uuid.Parse(clientID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid client ID"})
		return
	}
	/** Create log entry */
	payload := map[string]interface{}{
		"ogData":     ogData,
		"shortToken": shortToken,
	}

	b, _ := json.Marshal(payload)
	log := models.Log{
		UserID: uID,
		Action: "visit-link",
		Data:   datatypes.JSON(b),
	}
	/** Save log to database */
	if err := database.DB.Create(&log).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, link.URL)
}

func GetLinkStats(c *gin.Context) {
	shortToken := c.Param("shortToken")
	link, err := getLinkByShortToken(shortToken)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	var totalVisits int64
	var uniqueVisitors int64

	err = database.DB.Model(&models.Log{}).
		Where("action = ?", "visit-link").
		Where("data->>'shortToken' = ?", shortToken).
		Count(&totalVisits).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uniqueVisitors, err = countUniqueVisitors(shortToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"link":           link,
		"totalVisits":    totalVisits,
		"uniqueVisitors": uniqueVisitors,
	})
}

func DeleteLink(c *gin.Context) {
	shortToken := c.Param("shortToken")
	if err := deleteLink(shortToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Link deleted successfully"})
}
