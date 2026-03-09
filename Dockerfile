FROM golang:1.25-alpine AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

WORKDIR /build
COPY server/ .
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X 'main.buildTime=${BUILD_TIME}'" -o auxo-mcp-server .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /build/auxo-mcp-server /usr/local/bin/auxo-mcp-server

EXPOSE 8080

ENTRYPOINT ["auxo-mcp-server"]
CMD ["-mode", "HTTP"]
