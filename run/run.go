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
	Start := time.Now()

	Spec, TargetURL, UserName, Password, ApiKey, Token, OutputDir, Detailed, Duration, ExtraArgs, EnableDebug := cliArgs.ParseCliArgs()

	// Set log level based on the debug flag
	if EnableDebug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	Context, OpenapiData, Router := openapi.ParseOpenapiSpec(Spec, TargetURL)

	loginEndpoint := getLoginEndpoint(OpenapiData)

	// Try to create authentication cookies, but don't fail if authentication isn't available
	AuthCookies := auth.CreateUserWithLoginEndpoint(UserName, Password, TargetURL, loginEndpoint)
	log.Infof("Authentication cookies created: %d cookies", len(AuthCookies))

	if len(AuthCookies) == 0 {
		log.Warn("No authentication cookies obtained - the target may not require authentication or the login endpoint may not exist")
	}

	fmt.Printf("Starting fuzzing for duration: %v\n", Duration)
	fmt.Printf("Target URL: %s\n", TargetURL)
	fmt.Printf("Fuzzing started at: %s\n", Start.Format("2006-01-02 15:04:05"))

	// Log the number of paths in the OpenAPI spec
	pathCount := 0
	for path, pathItem := range OpenapiData.Paths.Map() {
		if pathItem != nil {
			operations := pathItem.Operations()
			pathCount += len(operations)
			log.Debugf("Found path: %s with %d operations", path, len(operations))
			for method := range operations {
				log.Debugf("  - Method: %s", method)
			}
		}
	}
	// Note: Total operations count is already logged in openapi module

	requestCount := 0
	log.Debugf("Starting fuzzing loop with duration: %v", Duration)
	for time.Since(Start) < Duration {
		log.Debugf("Starting request creation iteration at %v", time.Since(Start))
		Requests, RequestsValidationInput, RequestsValidationError := request.CreateRequest(*Context, OpenapiData, Router, TargetURL, AuthCookies, UserName, Password, ApiKey, Token, ExtraArgs, EnableDebug)
		log.Debugf("Created %d requests in this iteration", len(Requests))
		log.Debugf("Request validation inputs: %d", len(RequestsValidationInput))
		log.Debugf("Request validation errors: %d", len(RequestsValidationError))

		// Report progress
		requestCount += len(Requests)
		elapsed := time.Since(Start)
		fmt.Printf("Progress: %d requests sent in %v | Current time: %s\r",
			requestCount, elapsed, time.Now().Format("15:04:05"))

		if len(Requests) > 0 {
			log.Debugf("Sending %d requests to response.ParseResponse", len(Requests))
		} else {
			log.Debugf("No requests created in this iteration, checking why...")
			// Additional debugging if no requests are created
			log.Debugf("Target URL: %s", TargetURL)
			log.Debugf("Auth cookies count: %d", len(AuthCookies))
		}

		response.ParseResponse(*Context, Requests, RequestsValidationInput, RequestsValidationError, OutputDir, Detailed)
	}

	fmt.Printf("\nFuzzing completed. Total requests sent: %d\n", requestCount)
	fmt.Printf("Total duration: %v\n", time.Since(Start))
}

// getLoginEndpoint attempts to find the login endpoint from the OpenAPI spec
func getLoginEndpoint(spec *openapi3.T) string {
	// Look for common login endpoint patterns
	loginPatterns := []string{"/login", "/users/login", "/user/login", "/api/login", "/auth/login", "/users/v1/login", "/api/v3/user/login"}

	for path, pathItem := range spec.Paths.Map() {
		if pathItem == nil {
			continue
		}

		// Track if we've already logged this endpoint to avoid duplicates
		loggedEndpoint := false

		// Check if this path is a login endpoint based on the operation ID or path
		lowerPath := strings.ToLower(path)
		for _, pattern := range loginPatterns {
			if strings.Contains(lowerPath, pattern) {
				// Check if it's a POST operation (typical for login)
				if pathItem.Post != nil {
					log.Infof("Detected login endpoint: %s", path)
					loggedEndpoint = true
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
						// Only log if we haven't already logged this endpoint
						if !loggedEndpoint {
							log.Infof("Detected login endpoint from operation ID: %s", path)
						}
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
