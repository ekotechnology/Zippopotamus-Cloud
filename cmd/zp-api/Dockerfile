FROM golang:1.13.4 as builder
MAINTAINER Trea Hauet <trea@zippopotam.us>

ARG path=/go/src/github.com/zippopotamus/zippopotamus
ARG internalPath=github.com/zippopotamus/zippopotamus/internal
ARG version="0.0.0"
ARG gitCommit="NOSHA"
ARG buildDate="0000-00-00"

ARG ldFlags="-X '$internalPath.Version=$version' -X '$internalPath.GitCommit=$gitCommit' -X '$internalPath.BuildDate=$buildDate'"

WORKDIR $path
COPY go.mod go.sum $path/
COPY vendor $path/vendor
COPY internal $path/internal
COPY cmd/zp-api $path/cmd/zp-api

RUN cd cmd/zp-api && GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$ldFlags" -o /tmp/zp-api .

FROM scratch as api
WORKDIR /opt/zp
USER 9999:9999
COPY --chown=9999:9999 --from=builder /tmp/zp-api /usr/local/bin/zp-api
CMD ["zp-api"]

