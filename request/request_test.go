package request

import "testing"

func TestBuildOperationURLUsesSpecBasePathWithHostOnlyTarget(t *testing.T) {
	fullURL, fullPath, routePath, err := buildOperationURL("http://localhost:8080", "/api/v3", "/pet")
	if err != nil {
		t.Fatalf("buildOperationURL returned error: %v", err)
	}

	if fullURL != "http://localhost:8080/api/v3/pet" {
		t.Fatalf("fullURL = %q, want %q", fullURL, "http://localhost:8080/api/v3/pet")
	}
	if fullPath != "/api/v3/pet" {
		t.Fatalf("fullPath = %q, want %q", fullPath, "/api/v3/pet")
	}
	if routePath != "/api/v3/pet" {
		t.Fatalf("routePath = %q, want %q", routePath, "/api/v3/pet")
	}
}

func TestBuildOperationURLUsesTargetBasePathWhenProvided(t *testing.T) {
	fullURL, fullPath, routePath, err := buildOperationURL("http://localhost:8080/custom", "/custom", "/pet")
	if err != nil {
		t.Fatalf("buildOperationURL returned error: %v", err)
	}

	if fullURL != "http://localhost:8080/custom/pet" {
		t.Fatalf("fullURL = %q, want %q", fullURL, "http://localhost:8080/custom/pet")
	}
	if fullPath != "/custom/pet" {
		t.Fatalf("fullPath = %q, want %q", fullPath, "/custom/pet")
	}
	if routePath != "/custom/pet" {
		t.Fatalf("routePath = %q, want %q", routePath, "/custom/pet")
	}
}

func TestBuildOperationURLUsesAbsoluteServerWhenTargetIsEmpty(t *testing.T) {
	fullURL, fullPath, routePath, err := buildOperationURL("", "https://api.example.com/v1", "/pet")
	if err != nil {
		t.Fatalf("buildOperationURL returned error: %v", err)
	}

	if fullURL != "https://api.example.com/v1/pet" {
		t.Fatalf("fullURL = %q, want %q", fullURL, "https://api.example.com/v1/pet")
	}
	if fullPath != "/v1/pet" {
		t.Fatalf("fullPath = %q, want %q", fullPath, "/v1/pet")
	}
	if routePath != "/v1/pet" {
		t.Fatalf("routePath = %q, want %q", routePath, "/v1/pet")
	}
}

func TestBuildOperationURLRequiresTargetForRelativeServer(t *testing.T) {
	_, _, _, err := buildOperationURL("", "/api/v3", "/pet")
	if err == nil {
		t.Fatal("buildOperationURL returned nil error for relative server without target")
	}
}
