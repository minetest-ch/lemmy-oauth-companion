FROM alpine:3.20.3
COPY lemmy-oauth-companion /bin/lemmy-oauth-companion
EXPOSE 8080
ENTRYPOINT ["/bin/lemmy-oauth-companion"]