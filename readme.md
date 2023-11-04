
Lemmy sidecar/companion for oauth login

# How it works

**Warning**:
This project uses script injection via `LEMMY_UI_CUSTOM_HTML_HEADER` and some routing-tricks (see `nginx.conf`) to achieve its goal.

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
docker-compose up --build
```

Log in to http://localhost:8000 with username `admin` and password `enterenter`

# License

* MIT