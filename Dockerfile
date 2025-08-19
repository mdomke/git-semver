FROM golang:1.24-trixie as builder
WORKDIR /go/src/github.com/mdomke/git-semver
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-w -s -extldflags="-static"' -a -o /go/bin/git-semver

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /go/bin/git-semver /bin/git-semver
WORKDIR /git-semver
ENTRYPOINT ["/bin/git-semver"]
