package response

import (
	"context"
	"io"
	"net/http"
	"shelob/logging"
	"strings"

	"github.com/getkin/kin-openapi/openapi3filter"
	log "github.com/sirupsen/logrus"
)

func ParseResponse(ctx context.Context, httpRequests []*http.Request, requestsValidationInput []*openapi3filter.RequestValidationInput, requestsValidationError []*error, outputDir string, detailed bool) {
	logging.CreateDir(outputDir)
	log.Debugf("response.go	Total requests to send: %d", len(httpRequests))

	if len(httpRequests) == 0 {
		log.Debugf("response.go	No requests to send - this might be the issue!")
		return
	}

	for idx := range httpRequests {
		log.Debugf("response.go	About to send request: %s %s", httpRequests[idx].Method, httpRequests[idx].URL.String())
		func() {
			client := &http.Client{}
			// We need to get the body again for logging purposes, but don't consume it unnecessarily
			// Just get the body for logging if needed
			var requestBody []byte
			if httpRequests[idx].GetBody != nil {
				requestBodyCopy, err := httpRequests[idx].GetBody()
				if err != nil {
					log.Warn("response.go	Failed to create copy of body: ", err)
					// Continue anyway without the body
					requestBody = []byte{}
				} else {
					requestBody, err = io.ReadAll(requestBodyCopy)
					if err != nil {
						log.Error("response.go	Failed to read copy of body: ", err)
						// Continue anyway without the body
						requestBody = []byte{}
					}
				}
			} else {
				requestBody = []byte{}
			}

			log.Debugf("response.go	Sending request: %s %s", httpRequests[idx].Method, httpRequests[idx].URL.String())
			log.Debugf("response.go	Request URL: %s", httpRequests[idx].URL.String())
			log.Debugf("response.go	Request method: %s", httpRequests[idx].Method)
			log.Debugf("response.go	Request headers: %v", httpRequests[idx].Header)

			// Add more debugging to see if the request is actually being sent
			log.Debugf("response.go	About to make HTTP request to: %s", httpRequests[idx].URL.String())
			log.Debugf("response.go	Request method: %s", httpRequests[idx].Method)
			log.Debugf("response.go	Request headers: %v", httpRequests[idx].Header)
			log.Debugf("response.go	Request body length: %d", len(requestBody))

			httpResponse, err := client.Do(httpRequests[idx])
			if err != nil {
				log.Error("response.go	Failed to make http request: ", err)
				log.Debugf("response.go	Error details: %v", err)
				// Instead of returning, continue to next request to allow fuzzing to continue
				return
			}

			log.Debugf("response.go	HTTP request sent successfully, got response with status: %d", httpResponse.StatusCode)

			defer httpResponse.Body.Close()

			responseBody, err := io.ReadAll(httpResponse.Body)
			if err != nil {
				log.Error("response.go	Failed to read response body: ", err)
			}

			responseHeaders := httpRequests[idx].Header.Clone()
			responseCode := httpResponse.StatusCode

			log.Debugf("response.go	Response received: %d for %s %s", responseCode, httpRequests[idx].Method, httpRequests[idx].URL.String())

			err = ValidateResponse(ctx, requestsValidationInput[idx], responseCode, &responseHeaders)

			if err != nil {
				log.Debugf("response.go	Validation error for %s %s: %v", httpRequests[idx].Method, httpRequests[idx].URL.String(), err)
				// Create filename with host:port and path to better identify vulnerable sites
				hostPort := strings.ReplaceAll(httpRequests[idx].URL.Host, ":", "_") // Replace : with _ to avoid filesystem issues
				pathPart := strings.ReplaceAll(httpRequests[idx].URL.RawPath, "/", "_")
				if pathPart == "" || pathPart == "_" {
					pathPart = "_root" // Use _root for root path to make it more descriptive
				}
				logging.WrapCrash(outputDir+"/"+hostPort+pathPart, httpResponse, *requestsValidationError[idx], requestBody, responseBody, err)
			} else if detailed {
				log.Debugf("response.go	Validation passed for %s %s", httpRequests[idx].Method, httpRequests[idx].URL.String())
				// Create filename with host:port and path to better identify vulnerable sites
				hostPort := strings.ReplaceAll(httpRequests[idx].URL.Host, ":", "_") // Replace : with _ to avoid filesystem issues
				pathPart := strings.ReplaceAll(httpRequests[idx].URL.RawPath, "/", "_")
				if pathPart == "" || pathPart == "_" {
					pathPart = "_root" // Use _root for root path to make it more descriptive
				}
				logging.WrapTest(outputDir+"/"+hostPort+pathPart, httpResponse, *requestsValidationError[idx], requestBody, responseBody, err)
			}
		}()
	}
}

func ValidateResponse(ctx context.Context, requestValidationInput *openapi3filter.RequestValidationInput, responseCode int, responseHeaders *http.Header) error {
	responseBody := []byte(`{}`)
	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestValidationInput,
		Status:                 responseCode,
		Header:                 *responseHeaders,
		Options: &openapi3filter.Options{
			ExcludeResponseBody:   false,
			IncludeResponseStatus: true,
			MultiError:            true,
		},
	}
	responseValidationInput.SetBodyBytes(responseBody)
	err := openapi3filter.ValidateResponse(ctx, responseValidationInput)
	return err
}
