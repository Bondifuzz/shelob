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
	for idx := range httpRequests {
		func() {
			client := &http.Client{}
			requestBodyCopy, err := httpRequests[idx].GetBody()
			if err != nil {
				log.Warn("response.go	Failed to create copy of body: ", err)
			}

			requestBody, err := io.ReadAll(requestBodyCopy)
			if err != nil {
				log.Error("response.go	Failed to read copy of body: ", err)
			}

			httpResponse, err := client.Do(httpRequests[idx])
			if err != nil {
				log.Fatal("response.go	Failed to make http request: ", err)
			}

			defer httpResponse.Body.Close()

			responseBody, err := io.ReadAll(httpResponse.Body)
			if err != nil {
				log.Error("response.go	Failed to read response body: ", err)
			}

			responseHeaders := httpRequests[idx].Header.Clone()
			responseCode := httpResponse.StatusCode

			err = ValidateResponse(ctx, requestsValidationInput[idx], responseCode, &responseHeaders)

			if err != nil {
				logging.WrapCrash(outputDir+"/"+strings.ReplaceAll(httpRequests[idx].URL.RawPath, "/", "_"), httpResponse, *requestsValidationError[idx], requestBody, responseBody, err)
			} else if detailed {
				logging.WrapTest(outputDir+"/"+strings.ReplaceAll(httpRequests[idx].URL.RawPath, "/", "_"), httpResponse, *requestsValidationError[idx], requestBody, responseBody, err)
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
