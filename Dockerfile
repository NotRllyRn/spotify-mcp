FROM golang:1.27-alpine AS build
RUN apk add --no-cache ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN mkdir -p /out/data && chown 65532:65532 /out/data && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/spotify-mcp ./cmd/spotify-mcp

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build --chown=65532:65532 /out/spotify-mcp /spotify-mcp
COPY --from=build --chown=65532:65532 /out/data /data
USER 65532:65532
EXPOSE 8080 8888
ENTRYPOINT ["/spotify-mcp"]
CMD ["serve"]
