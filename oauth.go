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
)

func HandleOAuthRedirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p := oauth_providers[vars["provider"]]
	if p == nil {
		return
	}

	http.Redirect(w, r, p.GetRedirectURL(), http.StatusSeeOther)
}

func handleLogin(user *provider.OAuthUserInfo, password_marker string, w http.ResponseWriter) error {
	ctx := context.Background()

	lemmyclient, err := lemmy.New(os.Getenv("LEMMY_URL"))
	if err != nil {
		return fmt.Errorf("lemmy client error: %v", err)
	}

	normalized_username := strings.ReplaceAll(user.Name, "-", "_")
	normalized_username = strings.ReplaceAll(normalized_username, " ", "_")
	normalized_username = strings.ReplaceAll(normalized_username, ".", "_")

	search_res, err := lemmyclient.Search(ctx, lemmy.Search{
		Type: lemmy.NewOptional(lemmy.SearchTypeUsers),
		Q:    normalized_username,
	})
	if err != nil {
		return fmt.Errorf("search error: %v", err)
	}

	// check if there is already an account by that name
	var found_person *lemmy.Person
	for _, res := range search_res.Users {
		if res.Person.Name == normalized_username {
			found_person = &res.Person
			break
		}
	}

	if found_person == nil {
		// no person with that name, create one
		captcha, err := lemmyclient.Captcha(ctx)
		if err != nil {
			return fmt.Errorf("get captcha error: %v", err)
		}

		uuid := captcha.Ok.ValueOrZero().UUID
		answer, err := lemmydb.GetCaptchaAnswer(uuid)
		if err != nil {
			return fmt.Errorf("captcha answer error: %v", err)
		}

		mail := lemmy.NewOptionalNil[string]()
		if user.Email != "" {
			mail = lemmy.NewOptional(user.Email)
		}

		_, err = lemmyclient.Register(ctx, lemmy.Register{
			Email:          mail,
			Username:       normalized_username,
			Password:       password_marker,
			PasswordVerify: password_marker,
			CaptchaAnswer:  lemmy.NewOptional(answer),
			CaptchaUUID:    lemmy.NewOptional(uuid),
		})
		if err != nil {
			return fmt.Errorf("register error: %v", err)
		}
	}

	// log in with the password-marker for that oauth provider
	err = lemmyclient.ClientLogin(ctx, lemmy.Login{
		UsernameOrEmail: normalized_username,
		Password:        password_marker,
	})
	if err != nil {
		return fmt.Errorf("login error: %v (this can happen if you already have a registered account on another oauth provider)", err)
	}

	// sync user profile
	us := lemmy.SaveUserSettings{
		BotAccount:  lemmy.NewOptional(false),
		ShowAvatars: lemmy.NewOptional(true),
	}

	sync_account := false

	if user.AvatarURL != "" {
		sync_account = true
		us.Avatar = lemmy.NewOptional(user.AvatarURL)
	}

	// DisplayName needs to be at least 3 characters long
	if user.DisplayName != "" && len(user.DisplayName) >= 3 {
		sync_account = true
		us.DisplayName = lemmy.NewOptional(user.DisplayName)
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
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600 * 24 * 7,
	})

	// serve a "html-redirect" instead of a real 30x to work around this bug: https://stackoverflow.com/a/71467131
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
		<head>
			<meta http-equiv="refresh" content="1; url='/?%d'">
		</head>
		<body>
			<a href="/">Click here if you are not redirected automatically</a>
		</body>
	</html>
	`, time.Now().Unix())))

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
		w.Write([]byte(fmt.Sprintf("Get userinfo error: %v", err)))
		return
	}

	err = handleLogin(user, p.GetPasswordMarker(), w)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
}
