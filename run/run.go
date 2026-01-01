package run

import (
	"fmt"
	"shelob/auth"
	"shelob/cliArgs"
	"shelob/openapi"
	"shelob/request"
	"shelob/response"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
)

func Run() {
	// Set default log level to info to show debug messages for troubleshooting
	log.SetLevel(log.ErrorLevel)

	Start := time.Now()

	Spec, TargetURL, UserName, Password, ApiKey, Token, OutputDir, Detailed, Duration, ExtraArgs := cliArgs.ParseCliArgs()

	Context, OpenapiData, Router := openapi.ParseOpenapiSpec(Spec, TargetURL)

	loginEndpoint := getLoginEndpoint(OpenapiData)
	AuthCookies := auth.CreateUserWithLoginEndpoint(UserName, Password, TargetURL, loginEndpoint)

	fmt.Printf("Starting fuzzing for duration: %v\n", Duration)
	fmt.Printf("Target URL: %s\n", TargetURL)
	fmt.Printf("Fuzzing started at: %s\n", Start.Format("2006-01-02 15:04:05"))

	requestCount := 0
	for time.Since(Start) < Duration {
		Requests, RequestsValidationInput, RequestsValidationError := request.CreateRequest(*Context, OpenapiData, Router, TargetURL, AuthCookies, UserName, Password, ApiKey, Token, ExtraArgs)

		// Report progress
		requestCount += len(Requests)
		elapsed := time.Since(Start)
		fmt.Printf("Progress: %d requests sent in %v | Current time: %s\r",
			requestCount, elapsed, time.Now().Format("15:04:05"))

		response.ParseResponse(*Context, Requests, RequestsValidationInput, RequestsValidationError, OutputDir, Detailed)
	}

	fmt.Printf("\nFuzzing completed. Total requests sent: %d\n", requestCount)
	fmt.Printf("Total duration: %v\n", time.Since(Start))
}

// getLoginEndpoint attempts to find the login endpoint from the OpenAPI spec
func getLoginEndpoint(spec *openapi3.T) string {
	// Look for common login endpoint patterns
	loginPatterns := []string{"/login", "/users/login", "/user/login", "/api/login", "/auth/login", "/users/v1/login", "/api/v3/user/login"}

	for path, pathItem := range spec.Paths {
		if pathItem == nil {
			continue
		}

		// Check if this path is a login endpoint based on the operation ID or path
		lowerPath := strings.ToLower(path)
		for _, pattern := range loginPatterns {
			if strings.Contains(lowerPath, pattern) {
				// Check if it's a POST operation (typical for login)
				if pathItem.Post != nil {
					log.Infof("Detected login endpoint: %s", path)
					return path
				}
			}
		}

		// Also check operation IDs for login-related terms
		operations := pathItem.Operations()
		for method, operation := range operations {
			if operation != nil && operation.OperationID != "" {
				operationID := strings.ToLower(operation.OperationID)
				if strings.Contains(operationID, "login") || strings.Contains(operationID, "authenticate") {
					if method == "POST" { // Login operations are typically POST
						log.Infof("Detected login endpoint from operation ID: %s", path)
						return path
					}
				}
			}
		}
	}

	// Default fallback if no login endpoint is found
	log.Warn("No login endpoint detected in OpenAPI spec, using default: /api/v3/user/login")
	return "/api/v3/user/login"
}
