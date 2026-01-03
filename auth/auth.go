package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type UserItem struct {
	Username string `json:"username"`
	Password string `json:"password"`

	// Add any field which is required to do login operation

	// For example:
	// SessionMetadata string `json:"session_metadata"`
}


func CreateUser(username, password, url string) []*http.Cookie {
	return CreateUserWithLoginEndpoint(username, password, url, "/api/v3/user/login")
}

func CreateUserWithLoginEndpoint(username, password, url, loginEndpoint string) []*http.Cookie {
	testUser := UserItem{
		Username: username,
		Password: password,
	}

	return testUser.getCookies(url, loginEndpoint)
}

func (testUser *UserItem) getCookies(url, loginEndpoint string) []*http.Cookie {
	// Ensure URL has proper scheme
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	payload, err := json.Marshal(testUser)
	if err != nil {
		log.Error("auth.go	Failed to create json: ", err)
		return nil
	}

	bodyParams := bytes.NewReader(payload)

	// Use the provided login endpoint
	loginURL := url + loginEndpoint
	httpRequest, err := http.NewRequest("POST", loginURL, bodyParams)
	if err != nil {
		log.Error("auth.go	Failed to create http request: ", err)
		return nil
	}

	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("Content-type", "application/json")

	response, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		log.Warn("auth.go	Failed to make http request for authentication: ", err)
		// Return empty cookies instead of nil to allow fuzzing to continue without authentication
		return []*http.Cookie{}
	}

	defer response.Body.Close()

	if response.StatusCode == 200 {
		log.Info("[+++] Cookies are stored")
	} else {
		log.Warn("[---] No cookies ;(\t", response.Status)
		// Return empty cookies instead of nil to allow fuzzing to continue without authentication
		return []*http.Cookie{}
	}

	return response.Cookies()
}
