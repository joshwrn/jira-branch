package main

import (
	"github.com/joho/godotenv"
	"github.com/joshwrn/jira-branch/internal/app"
)

func main() {
	err := app.Init()
	if err != nil {
		app.Log.Fatal().Err(err).Msg("Error initializing logger")
	}

	err = godotenv.Load()
	if err != nil {
		app.Log.Fatal().Err(err).Msg("Error loading .env file")
	}
	app.Log.Info().Msg("Starting application")
	app.Run()
}
