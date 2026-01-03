package request

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"shelob/bodyParams"
	"shelob/security"
	"shelob/urlParams"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	log "github.com/sirupsen/logrus"
)

func CreateRequest(ctx context.Context, openapiData *openapi3.T, router *routers.Router, targetURL string, authCookies []*http.Cookie, username, password, apikey, token string, extraArgs []string) ([]*http.Request, []*openapi3filter.RequestValidationInput, []*error) {
	var (
		httpRequests            []*http.Request
		requestsValidationInput []*openapi3filter.RequestValidationInput
		requestsValidationError []*error
	)

	log.Debugf("request.go	Starting request creation for URL: %s", targetURL)
	log.Debugf("request.go	Number of auth cookies: %d", len(authCookies))

	// Get the base path from the OpenAPI spec
	var basePath string
	if openapiData.Servers != nil && len(openapiData.Servers) > 0 && openapiData.Servers[0] != nil {
		serverURL := openapiData.Servers[0].URL
		parsedURL, err := url.Parse(serverURL)
		if err != nil {
			log.Errorf("request.go	Failed to parse server URL %s: %v", serverURL, err)
			basePath = "/"
		} else {
			basePath = parsedURL.Path
			if basePath == "" {
				basePath = "/"
			}
		}
	} else {
		log.Warn("request.go	No servers defined in OpenAPI spec, using root path")
		basePath = "/"
	}

	// Ensure base path starts with /
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}

	// Ensure base path doesn't end with / unless it's just "/"
	if len(basePath) > 1 && strings.HasSuffix(basePath, "/") {
		basePath = strings.TrimSuffix(basePath, "/")
	}

	log.Debugf("request.go	Base path: %s", basePath)
	log.Debugf("request.go	Target URL: %s", targetURL)

	// Add global security feature here
	securityScheme := &openapiData.Components.SecuritySchemes

	// Skip logout operation for using the same cookies during fuzzing
	re := regexp.MustCompile("logout$")
	log.Debugf("request.go	Iterating through %d paths", len(openapiData.Paths.Map()))
	for path, pathItem := range openapiData.Paths.Map() {
		log.Debugf("request.go	Processing path: %s", path)
		if re.FindStringSubmatch(path) != nil {
			log.Warnf("request.go	Logout operation at %#v, skip it", path)
			continue
		}

		// Check if pathItem is nil
		if pathItem == nil {
			log.Debugf("request.go	Path item is nil for path: %s", path)
			continue
		}

		operations := pathItem.Operations()
		log.Debugf("request.go	Found %d operations for path: %s", len(operations), path)

		for method, operation := range operations {
			log.Debugf("request.go	Processing operation: %s %s", method, path)
			pathParams, queryParams, headerParams, cookieParams := urlParams.CreatePathParams(operation)
			log.Debugf("request.go	Path params: %v, Query params: %v", pathParams, queryParams)
			contentType, bodyPayload := bodyParams.CreateBodyData(operation)
			log.Debugf("request.go	Content type: %s, Body payload length: %d", contentType, bodyPayload.Len())
			security.CreateSecurityParams(operation, securityScheme, *queryParams, headerParams, cookieParams, username, password, apikey, token)

			// Construct the full URL
			// Avoid double slashes by ensuring basePath ends with / and path starts with /
			// but only one slash is present in the final path
			var fullPath string
			if basePath == "/" {
				fullPath = path
			} else {
				// Ensure basePath ends with / and path doesn't start with /
				cleanBasePath := strings.TrimSuffix(basePath, "/")
				cleanPath := strings.TrimPrefix(path, "/")
				if cleanPath == "" {
					fullPath = cleanBasePath
				} else {
					fullPath = cleanBasePath + "/" + cleanPath
				}
			}
			fullURL := targetURL + fullPath

			httpRequest, err := http.NewRequest(method, fullURL, bodyPayload)
			if err != nil {
				log.Error("request.go	Failed to create http request: ", err)
				continue
			}
			log.Debugf("request.go	Successfully created HTTP request: %s %s", method, fullURL)

			// Set path parameters
			httpRequest.URL.Path = urlParams.MapToString(fullPath, pathParams)
			httpRequest.URL.RawQuery = queryParams.Encode()

			// Set headers
			httpRequest.Header.Set("accept", contentType)
			for key, value := range headerParams {
				httpRequest.Header.Add(key, value)
			}

			// Set cookies
			for key, value := range cookieParams {
				httpRequest.Header.Add(key, value)
			}
			for idx := range authCookies {
				httpRequest.AddCookie(authCookies[idx])
			}

			if bodyPayload != nil {
				httpRequest.Header.Set("Content-Type", contentType)
			}

			// Find and validate route - use original path before parameter substitution
			originalRequest, err := http.NewRequest(method, httpRequest.URL.Scheme+"://"+httpRequest.URL.Host+fullPath, bodyPayload)
			if err != nil {
				log.Error("request.go	Failed to create original http request for routing: ", err)
				continue
			}

			// Copy headers and other properties for routing
			originalRequest.Header = make(http.Header)
			for k, v := range httpRequest.Header {
				originalRequest.Header[k] = v
			}

			log.Debugf("request.go	Attempting to find route for %s %s", method, fullPath)
			route, pathParamsVal, err := (*router).FindRoute(originalRequest)
			if err != nil {
				log.Debugf("request.go	Skipping route %s %s: %v", method, fullPath, err)
				continue
			}
			log.Debugf("request.go	Route found for %s %s", method, fullPath)

			// Don't skip based on validation errors - we want to fuzz even invalid requests
			requestValidationInput, _ := ValidateRequest(httpRequest, pathParamsVal, queryParams, route, ctx)
			log.Debugf("request.go	Validation completed for %s %s", method, fullPath)

			httpRequests = append(httpRequests, httpRequest)
			requestsValidationInput = append(requestsValidationInput, requestValidationInput)
			requestsValidationError = append(requestsValidationError, &err)
			log.Debugf("request.go	Appended request, total requests now: %d", len(httpRequests))
		}
	}

	log.Debugf("request.go	Total requests created: %d", len(httpRequests))

	return httpRequests, requestsValidationInput, requestsValidationError
}

func ValidateRequest(httpRequest *http.Request, pathParams map[string]string, queryParams *url.Values, route *routers.Route, ctx context.Context) (*openapi3filter.RequestValidationInput, error) {
	// Create authentication options
	options := &openapi3filter.Options{
		ExcludeRequestBody: false,
		MultiError:         true,
		AuthenticationFunc: func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
			// For now, we'll return nil to bypass authentication during fuzzing
			// In a real implementation, you'd validate the token properly
			return nil
		},
	}

	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:     httpRequest,
		PathParams:  pathParams,
		QueryParams: *queryParams,
		Route:       route,
		Options:     options,
	}

	// Try to validate the request, but don't fail if validation fails during fuzzing
	validationErr := openapi3filter.ValidateRequest(ctx, requestValidationInput)
	if validationErr != nil {
		// Log the validation error but continue anyway during fuzzing
		log.Debugf("request.go	Validation error: %v", validationErr)
		// Return the input even if validation failed, so requests can still be sent
		return requestValidationInput, nil
	}

	return requestValidationInput, nil
}
