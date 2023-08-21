package logging

import (
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

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

	file, err := os.OpenFile(filename+".json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)

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

	file, err := os.OpenFile(filename+".json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)

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
