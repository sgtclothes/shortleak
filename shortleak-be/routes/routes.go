package routes

import (
	"shortleak/controllers"
	"shortleak/middlewares"

	"github.com/gin-gonic/gin"
)

/** SetupRoutes initializes the routes for the application */
func SetupRoutes(r *gin.Engine) {
	/** Public routes */
	r.Use(middlewares.ClientIDMiddleware())
	{
		r.GET("/:shortToken", controllers.RedirectLink)
	}
	routes := r.Group("/api")
	auth := routes.Group("/auth")
	auth.Use()
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
		auth.POST("/logout", controllers.Logout)
	}
	link := routes.Group("/links")
	link.GET("/:shortToken", controllers.GetLinkByShortToken)
	link.Use(middlewares.AuthRequired())
	{
		link.GET("/user", controllers.GetLinksByUserAuth)
		link.DELETE("/:shortToken", controllers.DeleteLink)
	}
	r.Use(middlewares.AuthRequired())
	{
		r.POST("/shorten", controllers.CreateLink)
		r.GET("/stats/:shortToken", controllers.GetLinkStats)
	}
	// // Protected route
	// protected := r.Group("/api")
	// protected.Use(middlewares.AuthRequired())
	// {
	// 	protected.GET("/profile", func(c *gin.Context) {
	// 		user, _ := c.Get("user")
	// 		c.JSON(200, gin.H{"user": user})
	// 	})
	// }
}
