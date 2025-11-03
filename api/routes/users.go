package routes

import (
	"github.com/gin-gonic/gin"

	"goatsync/internal/handlers"
)

func setupUserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		users.POST("/login_challenge", handlers.LoginChallenge)
		users.POST("/login", handlers.Login)
		users.POST("/logout", handlers.Logout)
		users.POST("/change_password", handlers.ChangePassword)
		users.POST("/signup", handlers.SignUp)
		// users.POST("/dashboard_url", handlers.DashboardURL)
	}
}
