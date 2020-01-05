FROM golang:1.13 as build

WORKDIR /go/src/compote
ADD . /go/src/compote

RUN apt-get update && \
    apt-get -y install upx
RUN go get -d -v ./...
RUN go build -o /go/bin/compote -ldflags="-s -w" main.go
# RUN upx --ultra-brute compote

FROM gcr.io/distroless/base:nonroot
COPY --from=build --chown=nonroot:nonroot /go/bin/compote /
ENTRYPOINT ["/compote"]
