# Build from repo root:
#   docker build -f services/mcp/Dockerfile -t x-mcp .

FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY services/mcp/go.mod services/mcp/go.sum ./services/mcp/
RUN cd services/mcp && go mod download
COPY services/mcp/ ./services/mcp/
COPY platform.controlplane.json ./platform.controlplane.json
RUN cd services/mcp && CGO_ENABLED=0 go build -o /bin/mcp-server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/mcp-server /usr/local/bin/mcp-server
COPY --from=builder /src/platform.controlplane.json /app/platform.controlplane.json
VOLUME ["/root/.x-mcp"]
ENV MCP_PORT=8765
ENV MCP_ROOT=/app
EXPOSE 8765
CMD ["mcp-server"]
