FROM golang:1.13 as build

WORKDIR /go/src/compote
ADD . /go/src/compote

RUN apt-get update && \
    apt-get -y install upx && \
    go build -o /go/bin/compote -ldflags="-s -w" main.go && \
    echo "compressing..." && \
    upx --brute compote

FROM gcr.io/distroless/base:nonroot

LABEL org.opencontainers.image.authors="John Laswell" \
      org.opencontainers.image.url="https://github.com/jlaswell/compote" \
      org.opencontainers.image.source="https://github.com/jlaswell/compote" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.title="compote" \
      org.opencontainers.image.description="bite-sized dependency management for PHP"

WORKDIR /app
COPY --from=build --chown=nonroot:nonroot /go/bin/compote /
ENTRYPOINT ["/compote"]
CMD ["-f /app"]
