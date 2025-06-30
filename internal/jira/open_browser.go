package jira

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

var tempFileName = ""

func openBrowserWithTempFile(url string) error {
	tempFile, err := os.CreateTemp("", "oauth_*.html")
	if err != nil {
		return err
	}
	tempFileName = tempFile.Name()

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta http-equiv="refresh" content="0; url=%s"></head>
<body style="background-color: #000; color: #fff; text-align: center; padding: 20px;">
<p>Redirecting to authorization...</p>
</body></html>`, url)

	tempFile.WriteString(htmlContent)
	tempFile.Close()

	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd, args = "cmd", []string{"/c", "start", tempFile.Name()}
	case "darwin":
		cmd, args = "open", []string{tempFile.Name()}
	default:
		cmd, args = "xdg-open", []string{tempFile.Name()}
	}

	return exec.Command(cmd, args...).Start()
}
