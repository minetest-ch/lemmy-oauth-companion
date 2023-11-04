package provider

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type OAuthUserInfo struct {
	Name       string
	Email      string
	ExternalID string
}

type OAuthProvider interface {
	GetRedirectURL() string
	GetPasswordMarker() string
	GetUserInfo(code string) (*OAuthUserInfo, error)
}
