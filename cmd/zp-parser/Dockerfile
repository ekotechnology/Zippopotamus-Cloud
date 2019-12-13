FROM golang:1.13.4 as builder
MAINTAINER Trea Hauet <trea@zippopotam.us>

ARG path=/go/src/github.com/zippopotamus/zippopotamus
ARG internalPath=github.com/zippopotamus/zippopotamus/internal
ARG version="0.0.0-unset"
ARG gitCommit="NOSHA"
ARG buildDate="0000-00-00"

ARG ldFlags="-X '$internalPath.Version=$version' -X '$internalPath.GitCommit=$gitCommit' -X '$internalPath.BuildDate=$buildDate'"

WORKDIR $path
COPY go.mod go.sum $path/
COPY vendor $path/vendor
COPY internal $path/internal
COPY cmd/zp-parser $path/cmd/zp-parser

#RUN GO111MODULE=on go get -v all

RUN cd cmd/zp-parser && GO111MODULE=off CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$ldFlags" -o /tmp/zp-parser .

FROM scratch as parser
WORKDIR /opt/zp
COPY --from=builder /tmp/zp-parser /usr/local/bin/zp-parser
CMD ["zp-parser"]