FROM golang:1.11-alpine3.8 as builder
COPY . /go/src/github.com/mdomke/git-semver
RUN go install github.com/mdomke/git-semver

FROM alpine:3.8
COPY --from=builder /go/bin/git-semver /usr/local/bin/
ENTRYPOINT ["git-semver"]
