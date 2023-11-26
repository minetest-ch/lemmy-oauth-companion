package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/minetest-ch/lemmy-oauth-companion/provider"

	"github.com/gorilla/mux"
	"go.elara.ws/go-lemmy"
	"go.elara.ws/go-lemmy/types"
)

func HandleOAuthRedirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p := oauth_providers[vars["provider"]]
	if p == nil {
		return
	}

	http.Redirect(w, r, p.GetRedirectURL(), http.StatusSeeOther)
}

func handleLogin(user *provider.OAuthUserInfo, password_marker string, w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()

	lemmyclient, err := lemmy.New(os.Getenv("LEMMY_URL"))
	if err != nil {
		return fmt.Errorf("lemmy client error: %v", err)
	}

	normalized_username := strings.ReplaceAll(user.Name, "-", "_")
	normalized_username = strings.ReplaceAll(normalized_username, " ", "_")
	normalized_username = strings.ReplaceAll(normalized_username, ".", "_")

	search_res, err := lemmyclient.Search(ctx, types.Search{
		Type: types.NewOptional(types.SearchTypeUsers),
		Q:    normalized_username,
	})
	if err != nil {
		return fmt.Errorf("search error: %v", err)
	}

	// check if there is already an account by that name
	var found_person *types.Person
	for _, res := range search_res.Users {
		if res.Person.Name == normalized_username {
			found_person = &res.Person
			break
		}
	}

	if found_person == nil {
		// no person with that name, create one
		captcha, err := lemmyclient.Captcha(ctx, types.GetCaptcha{})
		if err != nil {
			return fmt.Errorf("get captcha error: %v", err)
		}

		uuid := captcha.Ok.MustValue().Uuid
		answer, err := lemmydb.GetCaptchaAnswer(uuid)
		if err != nil {
			return fmt.Errorf("captcha answer error: %v", err)
		}

		mail := types.NewOptionalNil[string]()
		if user.Email != "" {
			mail = types.NewOptional(user.Email)
		}

		_, err = lemmyclient.Register(ctx, types.Register{
			Email:          mail,
			Username:       normalized_username,
			Password:       password_marker,
			PasswordVerify: password_marker,
			CaptchaAnswer:  types.NewOptional(answer),
			CaptchaUuid:    types.NewOptional(uuid),
		})
		if err != nil {
			return fmt.Errorf("register error: %v", err)
		}
	}

	// log in with the password-marker for that oauth provider
	err = lemmyclient.ClientLogin(ctx, types.Login{
		UsernameOrEmail: normalized_username,
		Password:        password_marker,
	})
	if err != nil {
		return fmt.Errorf("login error: %v (this can happen if you already have a registered account on another oauth provider)", err)
	}

	// sync user profile
	us := types.SaveUserSettings{
		Auth:        lemmyclient.Token,
		BotAccount:  types.NewOptional(false),
		ShowAvatars: types.NewOptional(true),
	}

	sync_account := false

	if user.AvatarURL != "" {
		sync_account = true
		us.Avatar = types.NewOptional(user.AvatarURL)
	}

	// DisplayName needs to be at least 3 characters long
	if user.DisplayName != "" && len(user.DisplayName) >= 3 {
		sync_account = true
		us.DisplayName = types.NewOptional(user.DisplayName)
	}

	if sync_account {
		_, err = lemmyclient.SaveUserSettings(ctx, us)
		if err != nil {
			return fmt.Errorf("set user profile error: %v", err)
		}
	}

	// set the cookie with the returned jwt
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    lemmyclient.Token,
		Path:     "/",
		Secure:   os.Getenv("COOKIE_SECURE") == "true",
		Expires:  time.Now().Add(time.Hour * 24 * 7),
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	})

	// serve a "html-redirect" instead of a real 30x to work around this bug: https://stackoverflow.com/a/71467131
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
	<!DOCTYPE html>
	<html>
		<head>
			<meta http-equiv="refresh" content="0; url='/'">
		</head>
		<body>
			<a href="/">Click here if you are not redirected automatically</a>
		</body>
	</html>
	`))

	return nil
}

func HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p := oauth_providers[vars["provider"]]
	if p == nil {
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		// no code received, go back to main page
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	user, err := p.GetUserInfo(code)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	err = handleLogin(user, p.GetPasswordMarker(), w, r)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
}
