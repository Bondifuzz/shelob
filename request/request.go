package request

import (
	"context"
	"fmt"
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

func CreateRequest(ctx context.Context, openapiData *openapi3.T, router *routers.Router, targetURL string, authCookies []*http.Cookie, username, password, apikey, token string, extraArgs []string, debugEnabled bool) ([]*http.Request, []*openapi3filter.RequestValidationInput, []error) {
	var (
		httpRequests            []*http.Request
		requestsValidationInput []*openapi3filter.RequestValidationInput
		requestsValidationError []error
	)

	log.Debugf("request.go	Starting request creation for URL: %s", targetURL)
	log.Debugf("request.go	Number of auth cookies: %d", len(authCookies))

	serverURL := ""
	if openapiData.Servers != nil && len(openapiData.Servers) > 0 && openapiData.Servers[0] != nil {
		serverURL = openapiData.Servers[0].URL
	} else {
		log.Warn("request.go	No servers defined in OpenAPI spec, using root path")
	}

	log.Debugf("request.go	Server URL from spec: %s", serverURL)
	log.Debugf("request.go	Target URL: %s", targetURL)

	// Add global security feature here
	securityScheme := &openapiData.Components.SecuritySchemes

	// Skip logout operation for using the same cookies during fuzzing
	re := regexp.MustCompile("logout$")
	log.Debugf("request.go	Iterating through %d paths", len(openapiData.Paths.Map()))
	for path, pathItem := range openapiData.Paths.Map() {
		log.Debugf("request.go	Processing path: %s", path)
		if re.FindStringSubmatch(path) != nil {
			if debugEnabled {
				log.Warnf("request.go	Logout operation at %#v, skip it", path)
			}
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
			contentType, bodyPayload := bodyParams.CreateBodyData(operation, debugEnabled)
			log.Debugf("request.go	Content type: %s, Body payload length: %d", contentType, bodyPayload.Len())
			security.CreateSecurityParams(operation, securityScheme, *queryParams, headerParams, cookieParams, username, password, apikey, token)

			fullURL, fullPath, routePath, err := buildOperationURL(targetURL, serverURL, path)
			if err != nil {
				log.Errorf("request.go	Failed to build request URL for path %s: %v", path, err)
				continue
			}

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
			originalRequest, err := http.NewRequest(method, httpRequest.URL.Scheme+"://"+httpRequest.URL.Host+routePath, bodyPayload)
			if err != nil {
				log.Error("request.go	Failed to create original http request for routing: ", err)
				continue
			}

			// Copy headers and other properties for routing
			originalRequest.Header = make(http.Header)
			for k, v := range httpRequest.Header {
				originalRequest.Header[k] = v
			}

			log.Debugf("request.go	Attempting to find route for %s %s", method, routePath)
			route, pathParamsVal, err := (*router).FindRoute(originalRequest)
			if err != nil {
				log.Debugf("request.go	Skipping route %s %s: %v", method, routePath, err)
				continue
			}
			log.Debugf("request.go	Route found for %s %s", method, routePath)

			// Don't skip based on validation errors - we want to fuzz even invalid requests
			requestValidationInput, validationErr := ValidateRequest(httpRequest, pathParamsVal, queryParams, route, ctx)
			log.Debugf("request.go	Validation completed for %s %s", method, routePath)

			httpRequests = append(httpRequests, httpRequest)
			requestsValidationInput = append(requestsValidationInput, requestValidationInput)
			requestsValidationError = append(requestsValidationError, validationErr)
			log.Debugf("request.go	Appended request, total requests now: %d", len(httpRequests))
		}
	}

	log.Debugf("request.go	Total requests created: %d", len(httpRequests))

	return httpRequests, requestsValidationInput, requestsValidationError
}

func buildOperationURL(targetURL, serverURL, operationPath string) (string, string, string, error) {
	serverBasePath, err := pathFromServerURL(serverURL)
	if err != nil {
		return "", "", "", err
	}
	routePath := joinURLPath(serverBasePath, operationPath)

	if targetURL != "" {
		normalizedTargetURL := normalizeURL(targetURL)
		parsedTargetURL, err := url.Parse(normalizedTargetURL)
		if err != nil {
			return "", "", "", err
		}
		if parsedTargetURL.Scheme == "" || parsedTargetURL.Host == "" {
			return "", "", "", fmt.Errorf("target URL must include host: %s", targetURL)
		}

		targetBasePath := cleanURLPath(parsedTargetURL.Path)
		if targetBasePath == "/" {
			targetBasePath = serverBasePath
		}

		fullPath := joinURLPath(targetBasePath, operationPath)
		fullURL := parsedTargetURL.Scheme + "://" + parsedTargetURL.Host + fullPath
		return fullURL, fullPath, routePath, nil
	}

	if serverURL == "" {
		return "", "", "", fmt.Errorf("no server URL defined in spec and no target URL provided via CLI")
	}

	normalizedServerURL := normalizeURL(serverURL)
	parsedServerURL, err := url.Parse(normalizedServerURL)
	if err != nil {
		return "", "", "", err
	}
	if parsedServerURL.Scheme == "" || parsedServerURL.Host == "" {
		return "", "", "", fmt.Errorf("server URL is relative, provide -url for target host: %s", serverURL)
	}

	fullURL := parsedServerURL.Scheme + "://" + parsedServerURL.Host + routePath
	return fullURL, routePath, routePath, nil
}

func pathFromServerURL(serverURL string) (string, error) {
	if serverURL == "" {
		return "/", nil
	}

	parsedServerURL, err := url.Parse(serverURL)
	if err != nil {
		return "", err
	}

	return cleanURLPath(parsedServerURL.Path), nil
}

func cleanURLPath(path string) string {
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}

func joinURLPath(basePath, operationPath string) string {
	basePath = cleanURLPath(basePath)
	operationPath = cleanURLPath(operationPath)
	if basePath == "/" {
		return operationPath
	}
	if operationPath == "/" {
		return basePath
	}
	return basePath + "/" + strings.TrimPrefix(operationPath, "/")
}

func normalizeURL(rawURL string) string {
	if rawURL == "" || strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		return rawURL
	}
	return "http://" + rawURL
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
	if httpRequest.GetBody != nil {
		body, err := httpRequest.GetBody()
		if err != nil {
			log.Debugf("request.go	Failed to reset request body after validation: %v", err)
		} else {
			httpRequest.Body = body
		}
	}
	if validationErr != nil {
		// Log the validation error but continue anyway during fuzzing
		log.Debugf("request.go	Validation error: %v", validationErr)
		// Return the input even if validation failed, so requests can still be sent
		return requestValidationInput, validationErr
	}

	return requestValidationInput, nil
}
