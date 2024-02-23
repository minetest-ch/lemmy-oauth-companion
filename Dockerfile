FROM alpine:3.19.0
COPY lemmy-oauth-companion /bin/lemmy-oauth-companion
EXPOSE 8080
ENTRYPOINT ["/bin/lemmy-oauth-companion"]