package openapi

import (
	"context"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	log "github.com/sirupsen/logrus"
)

func ParseOpenapiSpec(spec string) (*context.Context, *openapi3.T, *routers.Router) {
	ctx := context.Background()
	loader := &openapi3.Loader{Context: ctx}
	openapiData, err := loader.LoadFromFile(spec)
	if err != nil {
		log.Fatal("openapi.go	Failed to load specification from file: ", err)
	}

	err = openapiData.Validate(ctx)

	if err != nil {
		log.Fatal("openapi.go	Failed to validate data: ", err)
	}

	router, err := gorillamux.NewRouter(openapiData)
	if err != nil {
		log.Fatal("openapi.go	Failed to create router: ", err)
	}

	log.Info("[+++] OpenAPI spec are parsed ok")

	return &ctx, openapiData, &router
}
