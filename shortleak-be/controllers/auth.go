package controllers

import (
	"net/http"
	"os"
	"shortleak/database"
	"shortleak/dto"
	"shortleak/models"
	"shortleak/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))
var generatePasswordHash = bcrypt.GenerateFromPassword
var signToken = func(token *jwt.Token, secret interface{}) (string, error) {
	return token.SignedString(secret)
}

/** Register a new user */
func Register(c *gin.Context) {
	var req dto.RegisterRequest

	/** Bind JSON to struct */
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	/** Validate request */
	if errors := utils.ValidateStruct(req); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"validation_errors": errors})
		return
	}

	/** Check if user already exists */
	var existingUser models.User
	if err := database.DB.First(&existingUser, "email = ?", req.Email).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		/** Hash the password */
		hashed, err := generatePasswordHash([]byte(req.Password), 10)
		if err != nil {
			return err
		}

		/** Create a new user */
		user := models.User{
			FullName: req.FullName,
			Email:    req.Email,
			Password: string(hashed),
		}

		/** Save user to database */
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		/** Create register log */
		log := models.Log{
			UserID: user.ID,
			Action: "register",
		}

		/** Save log to database */
		if err := tx.Create(&log).Error; err != nil {
			return err
		}

		return nil
	})

	/** Handle transaction error */
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to register user", "details": err.Error()})
		return
	}

	/** Return success */
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

/** Login user */
func Login(c *gin.Context) {
	platform := os.Getenv("PLATFORM")
	/** Validate request body */
	var req dto.LoginRequest

	/** Bind JSON to struct */
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	/** Validate request */
	if errors := utils.ValidateStruct(req); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"validation_errors": errors})
		return
	}

	/** Find user by email */
	var user models.User
	if err := database.DB.First(&user, "email = ?", req.Email).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	/** Compare password */
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	/** Create register log */
	log := models.Log{
		UserID: user.ID,
		Action: "login",
	}

	/** Save log to database */
	if err := database.DB.Create(&log).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log login attempt"})
		return
	}

	/** Create JWT token */
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(time.Hour * 24).Unix(), /** expired 1 day */
	})

	/** Sign the token with secret key */
	tokenString, err := signToken(token, jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	/** Set token in cookie */
	c.SetCookie(platform, tokenString, 3600*24, "/", "", false, true)

	var dataUser = map[string]interface{}{
		"id":       user.ID,
		"fullname": user.FullName,
		"email":    user.Email,
	}

	/** Return success */
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   tokenString,
		"user":    dataUser,
	})
}

/** Logout user */
func Logout(c *gin.Context) {
	/** Clear the token cookie */
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}
