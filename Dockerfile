FROM golang:1.11-alpine3.8 as builder
COPY . /go/src/github.com/mdomke/git-semver
RUN go install github.com/mdomke/git-semver

FROM docker:18-git
COPY --from=builder /go/bin/git-semver /usr/local/bin/
RUN mkdir /git-semver
WORKDIR /git-semver
ENTRYPOINT ["git-semver"]
