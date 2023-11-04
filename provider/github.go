package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type GithubOAuthProvider struct {
	ClientID     string
	ClientSecret string
}

type GithubAccessTokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
}

type GithubUserResponse struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

func (p *GithubOAuthProvider) GetRedirectURL() string {
	return fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&scope=user:email", p.ClientID)
}

func (p *GithubOAuthProvider) GetPasswordMarker() string {
	return p.ClientSecret
}

func (p *GithubOAuthProvider) GetUserInfo(code string) (*OAuthUserInfo, error) {
	accessTokenReq := GithubAccessTokenRequest{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Code:         code,
	}

	data, err := json.Marshal(accessTokenReq)
	if err != nil {
		return nil, fmt.Errorf("marshal access token error: %v", err)
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("new get token request error: %v", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get token error: %v", err)
	}
	defer resp.Body.Close()

	tokenData := AccessTokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(&tokenData)
	if err != nil {
		return nil, fmt.Errorf("token decode error: %v", err)
	}

	// fetch user data
	req, err = http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("new user request error: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenData.AccessToken))

	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get user error: %v", err)
	}
	defer resp.Body.Close()

	userData := GithubUserResponse{}
	err = json.NewDecoder(resp.Body).Decode(&userData)
	if err != nil {
		return nil, fmt.Errorf("user response error: %v", err)
	}

	oi := &OAuthUserInfo{
		Name:       userData.Login,
		Email:      userData.Email,
		AvatarURL:  fmt.Sprintf("https://github.com/%s.png", userData.Login),
		ExternalID: fmt.Sprintf("%d", userData.ID),
	}

	return oi, nil
}
