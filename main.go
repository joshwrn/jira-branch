package main

import (
	"github.com/joho/godotenv"
	"github.com/joshwrn/jira-branch/internal/app"
	"github.com/joshwrn/jira-branch/internal/utils"
)

func main() {
	err := utils.Init()
	if err != nil {
		utils.Log.Warn().Err(err).Msg("Error initializing logger")
	}

	err = godotenv.Load()
	if err != nil {
		utils.Log.Warn().Err(err).Msg("No .env file found, continuing with environment variables")
	}
	utils.Log.Info().Msg("Starting application")
	app.Run()
}
