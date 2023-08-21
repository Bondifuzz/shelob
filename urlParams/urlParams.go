package urlParams

import (
	"fmt"
	"net/url"
	"regexp"
	"shelob/generateInput"

	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
)

func CreatePathParams(operation *openapi3.Operation) (map[string]interface{}, *url.Values, map[string]interface{}, map[string]interface{}) {
	pathParams := make(map[string]interface{})
	headerParams := make(map[string]interface{})
	cookieParams := make(map[string]interface{})
	queryParams := &url.Values{}
	if operation.Parameters != nil {
		parameters := operation.Parameters
		for _, parameter := range parameters {
			if parameter.Value.Required {
				input := generateInput.GenerateRandomDataModels(parameter.Value.Schema.Value)
				switch parameter.Value.In {
				case "path":
					pathParams[parameter.Value.Name] = input
				case "query":
					queryParams.Add(parameter.Value.Name, input.(string))
				case "header":
					headerParams[parameter.Value.Name] = input
				case "cookie":
					cookieParams[parameter.Value.Name] = input
				default:
					log.Warn("urlParams.go	Unresolved parameter type: ", parameter.Value.In)
				}
			}
		}
	}
	return pathParams, queryParams, headerParams, cookieParams
}

func MapToString(path string, pathParams map[string]interface{}) string {
	for key, value := range pathParams {
		r := regexp.MustCompile("{" + key + "}")
		valueStr := fmt.Sprintf("%v", value)
		path = r.ReplaceAllString(path, valueStr)
	}
	return path
}
