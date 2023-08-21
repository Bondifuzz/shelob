package auth

import (
	"bytes"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type UserItem struct {
	Username string `json:"username"`
	Password string `json:"password"`

	// Add any field which is required to do login operation

	// For example:
	// SessionMetadata string `json:"session_metadata"`
}

// type Users struct {
//     user []UserItem
// }

func CreateUser(username, password, url string) []*http.Cookie {
	testUser := UserItem{
		Username: username,
		Password: password,
	}

	return testUser.getCookies(url)
}

func (testUser *UserItem) getCookies(url string) []*http.Cookie {
	payload, err := json.Marshal(testUser)
	if err != nil {
		log.Error("auth.go	Failed to create json: ", err)
	}

	bodyParams := bytes.NewReader(payload)

	// Correct path according to your login url

	httpRequest, err := http.NewRequest("POST", url+"/user/login", bodyParams)
	if err != nil {
		log.Error("auth.go	Failed to create http request: ", err)
	}

	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("Content-type", "application/json")

	response, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		log.Error("auth.go	Failed to make http request: ", err)
	}

	defer response.Body.Close()

	if response.StatusCode == 200 {
		log.Info("[+++] Cookies are stored")
	} else {
		log.Warn("[---] No cookies ;(\t", response.Status)
	}

	return response.Cookies()
}
