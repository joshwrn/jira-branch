package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/joshwrn/jira-branch/internal/app"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	app.Run()
}
