FROM golang:1.13.8-alpine3.10 as build

ENV GO111MODULE=on
ENV GOFLAGS="-mod=vendor"

RUN set -x \
  && apk add --update \
    g++ \
    gcc \
    curl

WORKDIR /go/src/gobak
COPY . /go/src/gobak

RUN go mod download
RUN go mod vendor

RUN go build