package middlewares

import (
	"net/http"
	"os"
	"shortleak/database"
	"shortleak/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

/** AuthRequired is a middleware to protect routes that require authentication */
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		/** Get token from cookie */
		tokenString, err := c.Cookie(os.Getenv("PLATFORM"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		/** Parse the token */
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		/** Check if token is valid */
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		/** Check if user exists */
		var user models.User
		if err := database.DB.First(&user, "id = ?", claims["userId"]).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		/** Set user in context */
		c.Set("user", user)
		c.Next()
	}
}

func ClientIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID, err := c.Cookie("client_id")
		if err != nil {
			/** Generate new client ID */
			clientID = uuid.New().String()
			c.SetCookie("client_id", clientID, 3600*24*365, "/", "", false, true)
		}
		c.Set("client_id", clientID)
		c.Next()
	}
}
