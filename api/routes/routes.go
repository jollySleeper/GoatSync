package routes

import (
	"github.com/gin-gonic/gin"
)

const VERSION = "v1"

// SetupRoutes configures all the routes for the application
func SetupRoutes(engine *gin.Engine) {
	// API version group
	apiEngine := engine.Group("/api/" + VERSION)

	// Setup route groups
	setupUserRoutes(apiEngine)
	setupMemberRoutes(apiEngine)
}
