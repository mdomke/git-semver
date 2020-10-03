FROM golang:1.14-alpine3.12 as builder
COPY . /go/src/github.com/mdomke/git-semver
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install github.com/mdomke/git-semver

FROM alpine:3.12
RUN apk add --update --no-cache git
COPY --from=builder /go/bin/git-semver /usr/local/bin/
WORKDIR /git-semver
ENTRYPOINT ["git-semver"]
