package openapi

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestNormalizeServersForTargetReplacesEmptyServerURL(t *testing.T) {
	doc := &openapi3.T{
		Servers: openapi3.Servers{&openapi3.Server{URL: ""}},
	}

	normalizeServersForTarget(doc, "http://127.0.0.1:5001")

	if len(doc.Servers) != 1 {
		t.Fatalf("servers length = %d, want 1", len(doc.Servers))
	}
	if doc.Servers[0].URL != "/" {
		t.Fatalf("server URL = %q, want %q", doc.Servers[0].URL, "/")
	}
}

func TestNormalizeServersForTargetKeepsSpecBasePath(t *testing.T) {
	doc := &openapi3.T{
		Servers: openapi3.Servers{&openapi3.Server{URL: "/api/v3"}},
	}

	normalizeServersForTarget(doc, "http://127.0.0.1:8080")

	if doc.Servers[0].URL != "/api/v3" {
		t.Fatalf("server URL = %q, want %q", doc.Servers[0].URL, "/api/v3")
	}
}

func TestNormalizeServersForTargetUsesCLIPathOverride(t *testing.T) {
	doc := &openapi3.T{
		Servers: openapi3.Servers{&openapi3.Server{URL: "/api/v3"}},
	}

	normalizeServersForTarget(doc, "http://127.0.0.1:8080/custom")

	if doc.Servers[0].URL != "/custom" {
		t.Fatalf("server URL = %q, want %q", doc.Servers[0].URL, "/custom")
	}
}
