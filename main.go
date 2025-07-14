package main

import (
	"github.com/joho/godotenv"
	"github.com/joshwrn/jira-branch/internal/app"
	"github.com/joshwrn/jira-branch/internal/utils"
)

func main() {
	envError := godotenv.Load()

	utils.Init()

	if envError != nil {
		utils.Log.Info().Err(envError).Msg("No .env file found, continuing with environment variables")
	}

	utils.Log.Info().Msg("Starting application")
	app.Run()
}
