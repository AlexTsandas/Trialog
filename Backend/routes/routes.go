package routes

import (
	"github.com/AlexTsandas/Trialog/Backend/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(app *gin.Engine) {
	app.GET("/test", handlers.GetTest)
}
