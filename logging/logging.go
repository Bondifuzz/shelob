package logging

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func sanitizeFilename(filename string) string {
	if filename == "" {
		return filename
	}

	// Clean the path to remove any relative path components like ../
	cleanPath := filepath.Clean(filename)

	// Get the base name to ensure we only use the filename part
	baseName := filepath.Base(cleanPath)

	// Additional check to ensure the base name doesn't contain path traversal
	// by using only the final component after any path separator
	if idx := max(strings.LastIndex(baseName, "/"), strings.LastIndex(baseName, "\\")); idx != -1 {
		baseName = baseName[idx+1:]
	}

	// If the result is "." (current directory), return an empty string or default name
	if baseName == "." {
		return ""
	}

	// Ensure the base name doesn't contain any path separators
	// This prevents path traversal by ensuring we only get the final component
	return baseName
}

func CreateDir(outputDir string) {
	err := os.MkdirAll(outputDir, 0o755)
	if err != nil && !os.IsExist(err) {
		log.Fatal("logging.go	Failed to create the directory: ", err)
	}
}

func WrapCrash(filename string, response *http.Response, requestValidationError error, requestBody []byte, responseBody []byte, errval error) {
	writeReport(filename, response, requestValidationError, requestBody, responseBody, errval, logrus.ErrorLevel, "Crash")
}

func WrapTest(filename string, response *http.Response, requestValidationError error, requestBody []byte, responseBody []byte, errval error) {
	writeReport(filename, response, requestValidationError, requestBody, responseBody, errval, logrus.InfoLevel, "Test")
}

func writeReport(filename string, response *http.Response, requestValidationError error, requestBody []byte, responseBody []byte, errval error, level logrus.Level, message string) {
	reportLog := logrus.New()
	reportLog.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint:       true,
		DisableHTMLEscape: true,
	})

	// Separate directory and base filename to preserve directory structure
	outputDir := filepath.Dir(filename)
	baseFilename := filepath.Base(filename)
	sanitizedBase := sanitizeFilename(baseFilename) // Only sanitize the base name
	if sanitizedBase == "" {
		sanitizedBase = "report"
	}

	timestamp := time.Now().Format("20060102_150405")
	file, err := os.CreateTemp(outputDir, sanitizedBase+"_"+timestamp+"_*.json")
	if err != nil {
		log.Debug("logging.go	Failed to log to the file: ", err)
		return
	}
	defer file.Close()
	reportLog.SetOutput(file)

	entry := reportLog.WithFields(logrus.Fields{
		"raw_path":            response.Request.URL.RawPath,
		"method":              response.Request.Method,
		"status":              response.Status,
		"path":                response.Request.URL.Path,
		"query":               response.Request.URL.RawQuery,
		"headers":             response.Request.Header,
		"response_headers":    response.Header,
		"cookies":             response.Request.Cookies(),
		"body_payload":        string(requestBody),
		"request_validation":  requestValidationError,
		"response_body":       string(responseBody),
		"response_validation": errval,
	})

	entry.Log(level, message)
}
