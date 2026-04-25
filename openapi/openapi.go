package openapi

import (
	"context"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	log "github.com/sirupsen/logrus"
)

func ParseOpenapiSpec(spec string, targetURL string) (*context.Context, *openapi3.T, *routers.Router) {
	ctx := context.Background()
	loader := &openapi3.Loader{Context: ctx}
	openapiData, err := loader.LoadFromFile(spec)
	if err != nil {
		log.Fatal("openapi.go	Failed to load specification from file: ", err)
	}

	if targetURL != "" {
		normalizeServersForTarget(openapiData, targetURL)
	} else {
		// Use server URLs from the spec file
		if openapiData.Servers != nil {
			log.Infof("Using server URLs from OpenAPI specification:")
			for i, server := range openapiData.Servers {
				if server != nil {
					log.Infof("  Server %d: %s", i+1, server.URL)
				}
			}
		} else {
			log.Warn("No servers defined in the OpenAPI specification and no target URL provided via CLI")
		}
	}

	err = openapiData.Validate(ctx)

	if err != nil {
		log.Fatal("openapi.go	Failed to validate data: ", err)
	}

	// Additional check: Ensure the spec has paths
	if openapiData.Paths == nil || len(openapiData.Paths.Map()) == 0 {
		log.Fatal("openapi.go	OpenAPI spec has no paths defined - invalid or empty spec file")
	}

	// Additional check: Ensure the spec has a proper structure
	if openapiData.Info == nil {
		log.Warn("openapi.go	OpenAPI spec has no info section")
	}

	// Count operations to ensure there are enough to fuzz
	totalOperations := 0
	log.Debugf("openapi.go	Analyzing paths in spec...")
	for path, pathItem := range openapiData.Paths.Map() {
		log.Debugf("openapi.go	Path: %s", path)
		if pathItem != nil {
			operations := pathItem.Operations()
			log.Debugf("openapi.go	Path %s has %d operations", path, len(operations))
			for method, operation := range operations {
				log.Debugf("openapi.go	- %s: %s", method, operation.Summary)
			}
			totalOperations += len(operations)
		} else {
			log.Debugf("openapi.go	Path %s has nil pathItem", path)
		}
	}

	log.Infof("openapi.go	Total operations in spec: %d", totalOperations)
	if totalOperations == 0 {
		log.Fatal("openapi.go	OpenAPI spec has paths but no operations defined - invalid spec file")
	}

	router, err := gorillamux.NewRouter(openapiData)
	if err != nil {
		log.Fatal("openapi.go	Failed to create router: ", err)
	}

	log.Info("[+++] OpenAPI spec are parsed ok")

	return &ctx, openapiData, &router
}

func normalizeServersForTarget(openapiData *openapi3.T, targetURL string) {
	targetPath := "/"
	parsedTargetURL, err := url.Parse(normalizeURL(targetURL))
	if err == nil && parsedTargetURL.Path != "" {
		targetPath = cleanServerPath(parsedTargetURL.Path)
	}

	// A CLI URL with a path is an explicit base-path override. A host-only CLI
	// URL keeps the base path from the OpenAPI servers section.
	if targetPath != "/" {
		log.Infof("Overriding server base paths in spec with CLI target URL path: %s", targetPath)
		openapiData.Servers = openapi3.Servers{
			&openapi3.Server{URL: targetPath},
		}
		return
	}

	if openapiData.Servers == nil || len(openapiData.Servers) == 0 {
		log.Infof("No servers defined in spec, using CLI target URL: %s", targetURL)
		openapiData.Servers = openapi3.Servers{
			&openapi3.Server{URL: "/"},
		}
		return
	}

	log.Infof("Using CLI target host with server base paths from the OpenAPI specification")
	for _, server := range openapiData.Servers {
		if server == nil {
			continue
		}

		serverPath := "/"
		if server.URL != "" {
			parsedServerURL, err := url.Parse(server.URL)
			if err == nil && parsedServerURL.Path != "" {
				serverPath = cleanServerPath(parsedServerURL.Path)
			}
		}
		server.URL = serverPath
	}
}

func normalizeURL(rawURL string) string {
	if rawURL == "" || strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		return rawURL
	}
	return "http://" + rawURL
}

func cleanServerPath(path string) string {
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
