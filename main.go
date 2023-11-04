package main

import (
	"database/sql"
	"embed"
	"fmt"
	"net/http"
	"os"

	"github.com/minetest-ch/lemmy-oauth-companion/provider"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var lemmydb *LemmyDatabase

var oauth_providers = map[string]provider.OAuthProvider{}

//go:embed assets/*
var Assets embed.FS

func main() {
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		panic(err)
	}

	lemmydb = &LemmyDatabase{db: db}

	oauth_providers["github"] = &provider.GithubOAuthProvider{
		ClientID:     os.Getenv("GITHUB_CLIENTID"),
		ClientSecret: os.Getenv("GITHUB_SECRET"),
		// BaseURL not needed, github uses the one configured in the oauth app
	}
	oauth_providers["cdb"] = &provider.CDBOAuthProvider{
		ClientID:     os.Getenv("CDB_CLIENTID"),
		ClientSecret: os.Getenv("CDB_SECRET"),
		BaseURL:      os.Getenv("BASE_URL"),
	}

	r := mux.NewRouter()
	r.HandleFunc("/oauth-login/{provider}", HandleOAuthRedirect)
	r.HandleFunc("/oauth-login/{provider}/callback", HandleOAuthCallback)

	r.PathPrefix("/oauth-login").Handler(http.StripPrefix("/oauth-login", http.FileServer(http.FS(Assets))))

	http.Handle("/", r)
	fmt.Println("Listening on port 8080")
	http.ListenAndServe(":8080", nil)
}
