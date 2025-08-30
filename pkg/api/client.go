package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var BackendURL = "http://localhost:8080"
var token string

func SetToken(t string) {
	token = t
}

func doRequest(req *http.Request) (*http.Response, error) {
	client := http.Client{Timeout: 8 * time.Second}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return client.Do(req)
}

func Health() (map[string]interface{}, error) {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(BackendURL + "/health")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	var m map[string]interface{}
	json.Unmarshal(b, &m)
	return m, nil
}

func Signup(email, password, master string) (*http.Response, error) {
	body := map[string]string{"email": email, "password": password, "master_password": master}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", BackendURL+"/api/v1/auth/signup", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return doRequest(req)
}

func Login(email, password, master string) (*http.Response, error) {
	body := map[string]string{"email": email, "password": password, "master_password": master}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", BackendURL+"/api/v1/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return doRequest(req)
}

// Secrets helpers
func GetSecrets(page, limit int) (*http.Response, error) {
	url := fmt.Sprintf("%s/api/v1/secrets?page=%d&limit=%d", BackendURL, page, limit)
	req, _ := http.NewRequest("GET", url, nil)
	return doRequest(req)
}

func CreateSecret(name, value, category, description, master string) (*http.Response, error) {
	body := map[string]string{"name": name, "value": value, "category": category, "description": description}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", BackendURL+"/api/v1/secrets", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Master-Password", master)
	return doRequest(req)
}

func UpdateSecret(id string, name, value, category, description, master string) (*http.Response, error) {
	body := map[string]string{"name": name, "value": value, "category": category, "description": description}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("PUT", BackendURL+"/api/v1/secrets/"+id, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Master-Password", master)
	return doRequest(req)
}

func DeleteSecret(id string) (*http.Response, error) {
	req, _ := http.NewRequest("DELETE", BackendURL+"/api/v1/secrets/"+id, nil)
	return doRequest(req)
}

// API Keys helpers
func GetAPIKeys(page, limit int, status string) (*http.Response, error) {
	url := fmt.Sprintf("%s/api/v1/apikeys?page=%d&limit=%d", BackendURL, page, limit)
	if status != "" {
		url = url + "&status=" + status
	}
	req, _ := http.NewRequest("GET", url, nil)
	return doRequest(req)
}

func CreateAPIKey(name string) (*http.Response, error) {
	body := map[string]string{"name": name}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", BackendURL+"/api/v1/apikeys", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return doRequest(req)
}

func RevokeAPIKey(id string) (*http.Response, error) {
	req, _ := http.NewRequest("POST", BackendURL+"/api/v1/apikeys/"+id+"/revoke", nil)
	return doRequest(req)
}

// GetCurrentUserEmail calls the backend /auth/me endpoint and returns the user's email if the current token (or API key) is valid.
func GetCurrentUserEmail() (string, error) {
	req, _ := http.NewRequest("GET", BackendURL+"/api/v1/auth/me", nil)
	resp, err := doRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("invalid response: %s", string(b))
	}
	var out map[string]interface{}
	json.Unmarshal(b, &out)
	if em, ok := out["email"].(string); ok {
		return em, nil
	}
	return "", fmt.Errorf("no email in response")
}

// ValidateAPIKey will attempt to validate the provided API key by temporarily setting it as the client token
// and calling the backend for the current user. If valid, token is kept and the user's email is returned.
func ValidateAPIKey(key string) (string, error) {
	prev := token
	token = key
	email, err := GetCurrentUserEmail()
	if err != nil {
		// restore previous token on failure
		token = prev
		return "", err
	}
	// keep token set to the validated key
	return email, nil
}

func HasToken() bool {
	return token != ""
}
