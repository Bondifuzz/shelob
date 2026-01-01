package security

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
)

// AuthenticationFunc is a function that validates OAuth2 tokens
type AuthenticationFunc func(ctx interface{}, input *openapi3filter.AuthenticationInput) error

// OAuth2Authenticator handles OAuth2 authentication
type OAuth2Authenticator struct {
	Token string
}

// Authenticate implements the AuthenticationFunc interface for OAuth2
func (a *OAuth2Authenticator) Authenticate(ctx interface{}, input *openapi3filter.AuthenticationInput) error {
	if a.Token == "" {
		return fmt.Errorf("missing OAuth2 token")
	}
	return nil
}

func CreateSecurityParams(operation *openapi3.Operation, securityScheme *openapi3.SecuritySchemes, queryParams url.Values, headerParams, cookieParams map[string]string, username, password, apikey, token string) {
	if operation.Security == nil {
		return
	}

	for _, securityItem := range *operation.Security {
		// First check for API key
		for schemeName := range securityItem {
			if scheme, ok := (*securityScheme)[schemeName]; ok {
				if scheme.Value.Type == "apiKey" {
					if apikey != "" {
						switch scheme.Value.In {
						case "query":
							queryParams.Set(scheme.Value.Name, apikey)
						case "header":
							headerParams[scheme.Value.Name] = apikey
						case "cookie":
							cookieParams[scheme.Value.Name] = apikey
						}
					}
				} else if scheme.Value.Type == "oauth2" {
					// Handle OAuth2 - add token if available
					if token != "" {
						headerParams["Authorization"] = "Bearer " + token
					}
				} else if scheme.Value.Type == "http" {
					if scheme.Value.Scheme == "basic" && username != "" && password != "" {
						auth := username + ":" + password
						headerParams["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
					} else if scheme.Value.Scheme == "bearer" && token != "" {
						headerParams["Authorization"] = "Bearer " + token
					}
				} else {
					// Log a debug message instead of warn for unsupported types during fuzzing
					// This prevents excessive warnings during fuzzing operations
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

// NewOAuth2Authenticator creates a new OAuth2 authenticator
func NewOAuth2Authenticator(token string) *OAuth2Authenticator {
	return &OAuth2Authenticator{
		Token: token,
	}
}

// ValidateOAuth2Token validates an OAuth2 token
func ValidateOAuth2Token(token string) bool {
	// Basic validation - check if token is not empty and has a valid format
	if token == "" {
		return false
	}

	// Check if token is in Bearer format
	parts := strings.Split(token, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return false
	}

	// Add more validation as needed
	return true
}
