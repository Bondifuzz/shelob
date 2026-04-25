package response

import (
	"context"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func TestValidateResponseUsesActualResponseBody(t *testing.T) {
	ctx := context.Background()
	operation := openapi3.NewOperation()
	responseSchema := openapi3.NewObjectSchema().
		WithProperty("name", openapi3.NewStringSchema()).
		WithRequired([]string{"name"})
	operation.AddResponse(200, openapi3.NewResponse().WithDescription("ok").WithJSONSchema(responseSchema))

	doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info:    &openapi3.Info{Title: "test", Version: "1.0.0"},
		Servers: openapi3.Servers{&openapi3.Server{URL: "http://example.com"}},
		Paths: openapi3.NewPaths(
			openapi3.WithPath("/pet", &openapi3.PathItem{Get: operation}),
		),
	}
	if err := doc.Validate(ctx); err != nil {
		t.Fatalf("spec validation failed: %v", err)
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		t.Fatalf("router creation failed: %v", err)
	}

	request, err := http.NewRequest(http.MethodGet, "http://example.com/pet", nil)
	if err != nil {
		t.Fatalf("request creation failed: %v", err)
	}
	route, pathParams, err := router.FindRoute(request)
	if err != nil {
		t.Fatalf("route lookup failed: %v", err)
	}

	validationInput := &openapi3filter.RequestValidationInput{
		Request:    request,
		PathParams: pathParams,
		Route:      route,
	}
	headers := http.Header{"Content-Type": []string{"application/json"}}

	if err := ValidateResponse(ctx, validationInput, http.StatusOK, headers, []byte(`{"name":"fluffy"}`)); err != nil {
		t.Fatalf("ValidateResponse returned error for valid body: %v", err)
	}

	if err := ValidateResponse(ctx, validationInput, http.StatusOK, headers, []byte(`{}`)); err == nil {
		t.Fatal("ValidateResponse returned nil error for invalid body")
	}
}
