FROM golang:1.16-buster as builder
ENV USER=semver
ENV UID=10001
RUN adduser \
 --disabled-password \
 --gecos "" \
 --home "/nonexistent" \
 --shell "/sbin/nologin" \
 --no-create-home \
 --uid "${UID}" \
 "${USER}"
WORKDIR /go/src/github.com/mdomke/git-semver
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-w -s -extldflags="-static"' -a -o /go/bin/git-semver

FROM scratch
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /go/bin/git-semver /bin/git-semver
USER semver:semver
WORKDIR /git-semver
ENTRYPOINT ["/bin/git-semver"]
