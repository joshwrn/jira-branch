package app

import (
	"encoding/json"
	"os"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func logObject(obj interface{}, msg string) {
	prettyJSON, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		Log.Error().Err(err).Msg("Failed to marshal object")
		return
	}

	Log.Info().Msgf("%s:\n%s", msg, string(prettyJSON))
}

func Init() error {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out:        logFile,
		TimeFormat: "15:04:05",
		NoColor:    false,
	}

	Log = zerolog.New(consoleWriter).With().Timestamp().Logger()
	return nil
}
