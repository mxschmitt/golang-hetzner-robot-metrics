FROM golang AS build-env

LABEL maintainer="Max Schmitt <max@schmitt.mx>"
LABEL description="Hetzner Robot Server Price metrics"

ADD . /go/src/github.com/mxschmitt/golang-hetzner-robot-metrics

RUN cd /go/src/github.com/mxschmitt/golang-hetzner-robot-metrics/cmd/metrics && \
    go get ./... && \
    CGO_ENABLED=0 go build -o /app

FROM alpine

RUN apk add ca-certificates

COPY --from=build-env /app /

ENTRYPOINT ["/app"]
