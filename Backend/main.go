package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/AlexTsandas/Trialog/Backend/config"
	"github.com/AlexTsandas/Trialog/Backend/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load("../.env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
	db, err := config.ConnectDB(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Database connection successful!")
	}
	defer db.Close()
	app := gin.Default()

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "5000"
	}
	routes.SetupRoutes(app)

	if err := app.Run(":" + PORT); err != nil {
		log.Fatal(err)
	}
}
