# vim:set ft=dockerfile:
FROM golang:1.12

COPY . /go/src/github.com/webitel/engine
WORKDIR /go/src/github.com/webitel/engine/

RUN GOOS=linux go get -d ./...
RUN GOOS=linux go install
RUN CGO_ENABLED=0 GOOS=linux go build -a -o engine .

FROM alpine:latest

LABEL maintainer="Vitaly Kovalyshyn"

ENV WEBITEL_MAJOR 19.12
ENV WEBITEL_REPO_BASE https://github.com/webitel

WORKDIR /
RUN apk --no-cache add ca-certificates tzdata
COPY --from=0 /go/src/github.com/webitel/engine/i18n /
COPY --from=0 /go/src/github.com/webitel/engine/engine .

ENTRYPOINT ["./engine"]
