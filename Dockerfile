# vim:set ft=dockerfile:
FROM golang:1.12

COPY . /go/src/github.com/webitel/engine
WORKDIR /go/src/github.com/webitel/engine/

RUN GOOS=linux go get -d ./...
RUN GOOS=linux go install
RUN GIT_COMMIT=$(git rev-list -1 HEAD) && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags "-X github.com/webitel/engine/model.BuildNumber=$GIT_COMMIT" -a -o engine .

FROM scratch

LABEL maintainer="Vitaly Kovalyshyn"

ENV WEBITEL_MAJOR 19.12
ENV WEBITEL_REPO_BASE https://github.com/webitel

WORKDIR /
COPY --from=0 /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /go/src/github.com/webitel/engine/i18n /i18n
COPY --from=0 /go/src/github.com/webitel/engine/engine /

ENTRYPOINT ["./engine"]
