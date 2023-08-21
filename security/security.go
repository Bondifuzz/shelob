package security

import (
	"encoding/base64"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
)

func CreateSecurityParams(operation *openapi3.Operation, securityScheme *openapi3.SecuritySchemes, queryParams *url.Values, headerParams, cookieParams map[string]interface{}, username, password, apikey, token string) {
	if operation.Security != nil {
		security := *operation.Security
		for _, items := range security {
			for key := range items {
				securityItemRef := findSecurityComponent(key, securityScheme)
				switch securityItemRef.Value.Type {
				case "apiKey":
					switch securityItemRef.Value.In {
					case "query":
						queryParams.Add(securityItemRef.Value.Name, apikey)
					case "header":
						headerParams[securityItemRef.Value.Name] = apikey
					case "cookie":
						cookieParams[securityItemRef.Value.Name] = apikey
					}
				case "http":
					switch securityItemRef.Value.Scheme {
					case "basic":
						basicAuth := []byte(username + ":" + password)
						headerParams["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString(basicAuth)
					case "bearer":
						headerParams["Authorization"] = "Bearer " + token
					}
				default:
					log.Warn("security.go	Unresolved security type: ", securityItemRef.Value.Type)
				}
			}
		}
	}
}

func findSecurityComponent(item string, securityScheme *openapi3.SecuritySchemes) *openapi3.SecuritySchemeRef {
	for name, schema := range *securityScheme {
		if item == name {
			return schema
		}
	}
	return nil
}
