package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/AlexTsandas/Trialog/Backend/config"
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

	fmt.Println("Hello, World!")

}
