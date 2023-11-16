package request

import (
	"context"
	"net/http"
	"net/url"
	"regexp"

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

	if basePath == "/" {
		basePath = ""
	}

	// Add global security feature here

	securityScheme := &openapiData.Components.SecuritySchemes

	// Skip logout operation for using the same cookies during fuzzing

	re := regexp.MustCompile("logout$")
	for path, pathItem := range openapiData.Paths {
		if re.FindStringSubmatch(path) != nil {
			log.Warnf("request.go	Logout operation at %#v, skip it", path)
		} else {
			for method, operation := range pathItem.Operations() {
				pathParams, queryParams, headerParams, cookieParams := urlParams.CreatePathParams(operation)
				contentType, bodyPayload := bodyParams.CreateBodyData(operation)
				security.CreateSecurityParams(operation, securityScheme, queryParams, headerParams, cookieParams, username, password, apikey, token)

				httpRequest, err := http.NewRequest(method, url+basePath+path, bodyPayload)
				if err != nil {
					log.Error("request.go	Failed to create http request: ", err)
				}

				httpRequest.URL.RawPath = basePath + path
				httpRequest.URL.Path = urlParams.MapToString(httpRequest.URL.Path, pathParams)
				httpRequest.URL.RawQuery = queryParams.Encode()

				httpRequest.Header.Set("accept", contentType)

				for key, value := range headerParams {
					httpRequest.Header.Add(key, value.(string))
				}

				for key, value := range cookieParams {
					httpRequest.Header.Add(key, value.(string))
				}

				if bodyPayload != nil {
					httpRequest.Header.Set("Content-Type", contentType)
				}

				for idx := range authCookies {
					httpRequest.AddCookie(authCookies[idx])
				}

				route, pathParamsVal, err := (*router).FindRoute(httpRequest)
				if err != nil {
					log.Error("request.go	Failed to find the applicable route: ", err)
				}

				requestValidationInput, err := ValidateRequest(httpRequest, pathParamsVal, queryParams, route, ctx)
				if err != nil {
					log.Error("request.go	Failed to validate request: ", err)
				}

				httpRequests = append(httpRequests, httpRequest)
				requestsValidationInput = append(requestsValidationInput, requestValidationInput)
				requestsValidationError = append(requestsValidationError, &err)
			}
		}
	}
	return httpRequests, requestsValidationInput, requestsValidationError
}

func ValidateRequest(httpRequest *http.Request, pathParams map[string]string, queryParams *url.Values, route *routers.Route, ctx context.Context) (*openapi3filter.RequestValidationInput, error) {
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:     httpRequest,
		PathParams:  pathParams,
		QueryParams: *queryParams,
		Route:       route,
		Options: &openapi3filter.Options{
			ExcludeRequestBody: false,
			MultiError:         true,
		},
	}
	err := openapi3filter.ValidateRequest(ctx, requestValidationInput)
	return requestValidationInput, err
}
