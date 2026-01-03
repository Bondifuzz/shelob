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
	log.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint:       true,
		DisableHTMLEscape: true,
	})

	// Create a unique filename with timestamp to avoid appending to the same file
	timestamp := time.Now().Format("20060102_150405_000") // Format: YYYYMMDD_HHMMSS_mmm

	// Separate directory and base filename to preserve directory structure
	outputDir := filepath.Dir(filename)
	baseFilename := filepath.Base(filename)
	sanitizedBase := sanitizeFilename(baseFilename) // Only sanitize the base name
	uniqueFilename := filepath.Join(outputDir, sanitizedBase + "_" + timestamp + ".json")

	file, err := os.OpenFile(uniqueFilename, os.O_CREATE|os.O_WRONLY, 0o644)

	if err == nil {
		log.SetOutput(file)
	} else {
		log.Warn("logging.go	Failed to log to the file, using default stderr")
	}

	// defer file.Close()

	log.WithFields(logrus.Fields{
		"raw_path":            response.Request.URL.RawPath,
		"method":              response.Request.Method,
		"status":              response.Status,
		"path":                response.Request.URL.Path,
		"query":               response.Request.URL.RawQuery,
		"headers":             response.Request.Header,
		"cookies":             response.Request.Cookies(),
		"body_payload":        string(requestBody),
		"request_validation":  requestValidationError,
		"response_body":       string(responseBody),
		"response_validation": errval,
	}).Error("Crash")
}

func WrapTest(filename string, response *http.Response, requestValidationError error, requestBody []byte, responseBody []byte, errval error) {
	log.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint:       true,
		DisableHTMLEscape: true,
	})

	// Create a unique filename with timestamp to avoid appending to the same file
	timestamp := time.Now().Format("20060102_150405_000") // Format: YYYYMMDD_HHMMSS_mmm

	// Separate directory and base filename to preserve directory structure
	outputDir := filepath.Dir(filename)
	baseFilename := filepath.Base(filename)
	sanitizedBase := sanitizeFilename(baseFilename) // Only sanitize the base name
	uniqueFilename := filepath.Join(outputDir, sanitizedBase + "_" + timestamp + ".json")

	file, err := os.OpenFile(uniqueFilename, os.O_CREATE|os.O_WRONLY, 0o644)

	if err == nil {
		log.SetOutput(file)
	} else {
		log.Warn("logging.go	Failed to log to the file, using default stderr")
	}

	// defer file.Close()

	log.WithFields(logrus.Fields{
		"raw_path":            response.Request.URL.RawPath,
		"method":              response.Request.Method,
		"status":              response.Status,
		"path":                response.Request.URL.Path,
		"query":               response.Request.URL.RawQuery,
		"headers":             response.Request.Header,
		"cookies":             response.Request.Cookies(),
		"body_payload":        string(requestBody),
		"request_validation":  requestValidationError,
		"response_body":       string(responseBody),
		"response_validation": errval,
	}).Info("Test")
}
