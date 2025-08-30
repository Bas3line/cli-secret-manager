package ui

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"sm-cli/pkg/api"
)

func (u *UI) ShowLogin() {
	// Prompt for API key only
	fields := []Field{{Label: "API Key", Width: 68}}
	vals, cancel := PromptForm(u.s, "Login with API Key", fields)
	if cancel {
		u.ShowMainMenu()
		return
	}

	key := vals["API Key"]
	if key == "" {
		DrawStatus(u.s, "API key required")
		u.ShowMainMenu()
		return
	}

	email, err := api.ValidateAPIKey(key)
	if err != nil {
		DrawStatus(u.s, fmt.Sprintf("Login failed: %v", err))
		u.ShowMainMenu()
		return
	}

	DrawStatus(u.s, "Login successful: "+email)
	// return to main menu where email will be shown
	u.ShowMainMenu()
}

func (u *UI) ShowSignup() {
	fields := []Field{{Label: "Email", Width: 40}, {Label: "Password", Width: 40, Masked: true}, {Label: "MasterPassword", Width: 40, Masked: true}}
	vals, cancel := PromptForm(u.s, "Signup", fields)
	if cancel {
		u.ShowMainMenu()
		return
	}

	resp, err := api.Signup(vals["Email"], vals["Password"], vals["MasterPassword"])
	if err != nil {
		DrawStatus(u.s, fmt.Sprintf("Signup failed: %v", err))
		u.ShowMainMenu()
		return
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 201 {
		DrawStatus(u.s, fmt.Sprintf("Signup failed: %s", string(b)))
		u.ShowMainMenu()
		return
	}
	var out map[string]interface{}
	json.Unmarshal(b, &out)
	if token, ok := out["token"].(string); ok {
		api.SetToken(token)
		DrawStatus(u.s, "Signup successful")
		u.ShowSecretsList(1)
		return
	}
	DrawStatus(u.s, "Signup failed: invalid server response")
	u.ShowMainMenu()
}

func (u *UI) ShowSecretsList(page int) {
	resp, err := api.GetSecrets(page, 20)
	if err != nil {
		DrawStatus(u.s, fmt.Sprintf("Failed to load secrets: %v", err))
		return
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		DrawStatus(u.s, fmt.Sprintf("Failed to load secrets: %s", string(b)))
		return
	}
	var out map[string]interface{}
	json.Unmarshal(b, &out)
	items := []string{}
	if arr, ok := out["secrets"].([]interface{}); ok {
		for i, v := range arr {
			if m, ok := v.(map[string]interface{}); ok {
				items = append(items, strconv.Itoa(i+1)+". "+fmt.Sprintf("%v", m["name"]))
			}
		}
	}
	if len(items) == 0 {
		items = append(items, "(no secrets)")
	}
	DrawList(u.s, items, 0)
	DrawStatus(u.s, "Loaded secrets")
}

func (u *UI) ShowAPIKeys(page int) {
	resp, err := api.GetAPIKeys(page, 20, "")
	if err != nil {
		DrawStatus(u.s, fmt.Sprintf("Failed to load apikeys: %v", err))
		return
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		DrawStatus(u.s, fmt.Sprintf("Failed to load apikeys: %s", string(b)))
		return
	}
	var out map[string]interface{}
	json.Unmarshal(b, &out)
	items := []string{}
	if arr, ok := out["api_keys"].([]interface{}); ok {
		for i, v := range arr {
			if m, ok := v.(map[string]interface{}); ok {
				items = append(items, strconv.Itoa(i+1)+". "+fmt.Sprintf("%v", m["name"]))
			}
		}
	}
	if len(items) == 0 {
		items = append(items, "(no api keys)")
	}
	DrawList(u.s, items, 0)
	DrawStatus(u.s, "Loaded api keys")
}
