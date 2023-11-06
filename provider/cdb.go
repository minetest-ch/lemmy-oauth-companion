package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
)

type CDBOAuthProvider struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
}

type CDBbUserResponse struct {
	Username string `json:"username"`
}

type CDBUser struct {
	Username      string `json:"username"`
	DisplayName   string `json:"display_name"`
	ProfilePicURL string `json:"profile_pic_url"` // "/uploads/xyz.jpg"
}

func (p *CDBOAuthProvider) GetRedirectURL() string {
	callback_url := fmt.Sprintf("%s/oauth-login/cdb/callback", p.BaseURL)
	return fmt.Sprintf("https://content.minetest.net/oauth/authorize/?response_type=code&client_id=%s&redirect_uri=%s", p.ClientID, url.QueryEscape(callback_url))
}

func (p *CDBOAuthProvider) GetPasswordMarker() string {
	return p.ClientSecret
}

func (p *CDBOAuthProvider) GetUserInfo(code string) (*OAuthUserInfo, error) {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("grant_type", "authorization_code")
	writer.WriteField("client_id", p.ClientID)
	writer.WriteField("client_secret", p.ClientSecret)
	writer.WriteField("code", code)
	err := writer.Close()
	if err != nil {
		return nil, fmt.Errorf("mulitpart error: %v", err)
	}

	req, err := http.NewRequest("POST", "https://content.minetest.net/oauth/token/", body)
	if err != nil {
		return nil, fmt.Errorf("new get token request error: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get token error: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status-code from token api: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	tokenData := AccessTokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(&tokenData)
	if err != nil {
		return nil, fmt.Errorf("token decode error: %v", err)
	}

	// fetch user data
	req, err = http.NewRequest("GET", "https://content.minetest.net/api/whoami/", nil)
	if err != nil {
		return nil, fmt.Errorf("new user request error: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenData.AccessToken))

	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get user error: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status-code from whoami api: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	userData := CDBbUserResponse{}
	err = json.NewDecoder(resp.Body).Decode(&userData)
	if err != nil {
		return nil, fmt.Errorf("user response error: %v", err)
	}

	if userData.Username == "" {
		return nil, fmt.Errorf("empty username from cdb received")
	}

	oi := &OAuthUserInfo{
		Name: userData.Username,
	}

	// fetch user profile
	req, err = http.NewRequest("GET", fmt.Sprintf("https://content.minetest.net/api/users/%s", userData.Username), nil)
	if err != nil {
		return nil, fmt.Errorf("new user-profile request error: %v", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get user-profile error: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status-code from user-profile api: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	userProfile := CDBUser{}
	err = json.NewDecoder(resp.Body).Decode(&userProfile)
	if err != nil {
		return nil, fmt.Errorf("user-profile response error: %v", err)
	}

	if userProfile.ProfilePicURL != "" {
		oi.AvatarURL = fmt.Sprintf("https://content.minetest.net%s", userProfile.ProfilePicURL)
	}
	oi.DisplayName = userProfile.DisplayName

	return oi, nil
}
