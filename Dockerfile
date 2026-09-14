FROM golang:alpine AS build-stage

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=v0.0.0-devel
WORKDIR /go/src/app

COPY . .

RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -v -a \
    -ldflags="-X 'github.com/pgvillage-tools/pgroute66/internal/version.Version=$VERSION'" -o pgroute66 ./cmd/pgroute66

FROM alpine AS export-stage
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
COPY --from=build-stage /go/src/app/pgroute66 /usr/bin/
COPY config/pgroute66.yaml /etc/pgroute66/config.yaml
CMD /usr/bin/pgroute66
