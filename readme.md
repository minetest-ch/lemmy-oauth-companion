
Lemmy sidecar/companion for oauth login

# Supported OAuth providers

* [Github](https://github.com)
* [Contentdb](https://content.minetest.ch)

# How it works

**Warning**:
This project uses script injection via `LEMMY_UI_CUSTOM_HTML_HEADER` and some routing-tricks (see `nginx.conf`) to achieve its goal.

Connections:
* pgsql connection to the lemmy database (for retrieving the captcha answer on signup)
* rest connection to the lemmy instance for user-search, signup and login

## Login process

* User clicks the "Login with xxx" button on the Login page
* User lands on the `/oauth-login/${oauth-provider}` page and gets redirected to the provider-login
* User gets redirected from the oauth-provider to the callback url `/oauth-login/${oauth-provider}/callback` with a code
* Backend retrieves the user-infos from the oauth-provider and creates a normalized/sanitized username
* If the username doesn't exist already: register new user via lemmy rest api (and captcha-answer from db)
* Log-in with the username and a per-provider password marker (to avoid account-takeover from other providers)
* Set jwt-cookie with value returned from login-call
* Redirect user to lemmy instance on `/`

# Dev

Setting up a dev-environment

Create an `.env` file with the configuration:
```
GITHUB_CLIENTID=
GITHUB_SECRET=
CDB_CLIENTID=
CDB_SECRET=
```

**Note**: callback url's are in the form of `http://localhost:8080/oauth-login/{provider}/callback` (provider is "github"/"cdb")

```sh
docker-compose up
```

Log in to http://localhost:8000 with username `admin` and password `enterenter`

# License

* MIT

## Other assets

* `assets/default_mese_crystal.png` CC BY-SA 3.0 https://github.com/minetest/minetest_game
* `assets/contentdb.png` GPL v3 https://github.com/minetest/contentdb