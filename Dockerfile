FROM --platform=${BUILDPLATFORM} golang:alpine AS gobuilder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -buildvcs=false \
    -ldflags="-w -s -extldflags '-static'" -o ./server ./cmd/server

FROM scratch
LABEL maintainer="wn@neessen.dev"
COPY ["docker-files/passwd", "/etc/passwd"]
COPY ["docker-files/group", "/etc/group"]
COPY --from=gobuilder ["/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/cert.pem"]

RUN mkdir /app && chown 1000:1000 /app && mkdir /app/data && chown 1000:1000 /app/data
WORKDIR /app
COPY etc/app.toml ./etc/app.toml
COPY --from=gobuilder /app/server ./

EXPOSE 10001
USER 1000

ENTRYPOINT ["./server"]
