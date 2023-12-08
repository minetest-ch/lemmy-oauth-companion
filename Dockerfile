FROM golang:1.21.4 as go-app
WORKDIR /data
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go test ./... -vet=off && CGO_ENABLED=0 go build .

FROM alpine:3.19.0
WORKDIR /
COPY --from=go-app /data/lemmy-oauth-companion /
CMD ["/lemmy-oauth-companion"]
