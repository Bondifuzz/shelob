package logging

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestWrapCrashCreatesUniqueFilesForRepeatedReports(t *testing.T) {
	outputDir := t.TempDir()
	request := httptest.NewRequest(http.MethodGet, "http://example.com/api/v3/user/login", nil)
	response := &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Request:    request,
	}

	filename := filepath.Join(outputDir, "example_api_v3_user_login")
	WrapCrash(filename, response, nil, nil, []byte(`{}`), assertErr("first"))
	WrapCrash(filename, response, nil, nil, []byte(`{}`), assertErr("second"))

	files, err := filepath.Glob(filepath.Join(outputDir, "example_api_v3_user_login_*.json"))
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("created %d report files, want 2", len(files))
	}
}

type assertErr string

func (err assertErr) Error() string {
	return string(err)
}
