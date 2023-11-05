package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type MesehubOAuthProvider struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
}

type MesehubUserResponse struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

func (p *MesehubOAuthProvider) GetRedirectURL() string {
	callback_url := fmt.Sprintf("%s/oauth-login/mesehub/callback", p.BaseURL)

	return fmt.Sprintf("https://git.minetest.land/login/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=STATE&scope=email",
		p.ClientID, callback_url)
}

func (p *MesehubOAuthProvider) GetPasswordMarker() string {
	return p.ClientSecret
}

func (p *MesehubOAuthProvider) GetUserInfo(code string) (*OAuthUserInfo, error) {
	accessTokenReq := make(map[string]string)
	accessTokenReq["client_id"] = p.ClientID
	accessTokenReq["client_secret"] = p.ClientSecret
	accessTokenReq["code"] = code
	accessTokenReq["grant_type"] = "authorization_code"
	accessTokenReq["redirect_uri"] = fmt.Sprintf("%s/oauth-login/mesehub/callback", p.BaseURL)

	data, err := json.Marshal(accessTokenReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://git.minetest.land/login/oauth/access_token", bytes.NewBuffer(data))
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
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code in token-response: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	tokenData := AccessTokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(&tokenData)
	if err != nil {
		return nil, fmt.Errorf("token decode error: %v", err)
	}

	// fetch user data
	req, err = http.NewRequest("GET", "https://git.minetest.land/api/v1/user", nil)
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

	userData := MesehubUserResponse{}
	err = json.NewDecoder(resp.Body).Decode(&userData)
	if err != nil {
		return nil, fmt.Errorf("user response error: %v", err)
	}

	oi := &OAuthUserInfo{
		Name:       userData.Login,
		Email:      userData.Email,
		AvatarURL:  fmt.Sprintf("https://git.minetest.land/%s.png", userData.Login),
		ExternalID: fmt.Sprintf("%d", userData.ID),
	}

	return oi, nil
}
