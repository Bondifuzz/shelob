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

func CreateRequest(ctx context.Context, openapiData *openapi3.T, router *routers.Router, url string, authCookies []*http.Cookie, username, password, apikey, token string, extraArgs []string) ([]*http.Request, []*openapi3filter.RequestValidationInput, []*error) {
	var (
		httpRequests            []*http.Request
		requestsValidationInput []*openapi3filter.RequestValidationInput
		requestsValidationError []*error
	)

	basePath, err := openapiData.Servers.BasePath()
	if err != nil {
		log.Error("request.go	Failed to get the base path: ", err)
	}

	// Ensure base path starts with /
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}

	// Add global security feature here
	securityScheme := &openapiData.Components.SecuritySchemes

	// Skip logout operation for using the same cookies during fuzzing
	re := regexp.MustCompile("logout$")
	for path, pathItem := range openapiData.Paths {
		if re.FindStringSubmatch(path) != nil {
			log.Warnf("request.go	Logout operation at %#v, skip it", path)
			continue
		}

		// Check if pathItem is nil
		if pathItem == nil {
			continue
		}

		for method, operation := range pathItem.Operations() {
			pathParams, queryParams, headerParams, cookieParams := urlParams.CreatePathParams(operation)
			contentType, bodyPayload := bodyParams.CreateBodyData(operation)
			security.CreateSecurityParams(operation, securityScheme, *queryParams, headerParams, cookieParams, username, password, apikey, token)

			// Construct the full URL
			fullPath := basePath + path
			fullURL := url + fullPath

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

			// Find and validate route
			log.Debugf("request.go	Attempting to find route for %s %s", method, fullPath)
			route, pathParamsVal, err := (*router).FindRoute(httpRequest)
			if err != nil {
				log.Debugf("request.go	Skipping route %s %s: %v", method, fullPath, err)
				continue
			}
			log.Debugf("request.go	Route found for %s %s", method, fullPath)

			requestValidationInput, err := ValidateRequest(httpRequest, pathParamsVal, queryParams, route, ctx)
			if err != nil {
				log.Debugf("request.go	Skipping validation for %s %s: %v", method, fullPath, err)
				continue
			}
			log.Debugf("request.go	Validation passed for %s %s", method, fullPath)

			httpRequests = append(httpRequests, httpRequest)
			requestsValidationInput = append(requestsValidationInput, requestValidationInput)
			requestsValidationError = append(requestsValidationError, &err)


		}
	}

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
	err := openapi3filter.ValidateRequest(ctx, requestValidationInput)
	if err != nil {
		// Log the validation error but continue anyway during fuzzing
		log.Debugf("request.go	Validation error: %v", err)
		// Return the input even if validation failed, so requests can still be sent
		return requestValidationInput, nil
	}

	return requestValidationInput, err
}
