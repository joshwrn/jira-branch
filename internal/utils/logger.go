package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func LogObject(obj interface{}, msg string) {
	prettyJSON, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		Log.Error().Err(err).Msg("Failed to marshal object")
		return
	}

	Log.Info().Msgf("%s:\n%s", msg, string(prettyJSON))
}

func getLogFilePath() (string, error) {
	if os.Getenv("DEV") == "true" {
		return "app.log", nil
	}

	var baseDir string
	var err error

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("LOCALAPPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Local")
		}
		baseDir = filepath.Join(appData, "jira-branch")
	case "darwin":
		baseDir = filepath.Join(homeDir, "Library", "Application Support", "jira-branch")
	default:
		xdgDataHome := os.Getenv("XDG_DATA_HOME")
		if xdgDataHome == "" {
			xdgDataHome = filepath.Join(homeDir, ".local", "share")
		}
		baseDir = filepath.Join(xdgDataHome, "jira-branch")
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(baseDir, "app.log"), nil
}

func Init() error {
	logFilePath, err := getLogFilePath()
	if err != nil {
		return err
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out:        logFile,
		TimeFormat: "15:04:05",
		NoColor:    false,
	}

	Log = zerolog.New(consoleWriter).With().Timestamp().Logger()

	if os.Getenv("DEV") == "true" {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	return nil
}
