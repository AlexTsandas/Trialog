package routes

import (
	"github.com/AlexTsandas/Trialog/Backend/handlers"
	"github.com/AlexTsandas/Trialog/Backend/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(app *gin.Engine) {
	app.GET("/test", middleware.RequireAuth, handlers.GetTest)
	app.POST("/auth/signup", handlers.SignUp)
	app.POST("/auth/login", handlers.SignIn)
	app.POST("/auth/refresh", handlers.Refresh)
}
