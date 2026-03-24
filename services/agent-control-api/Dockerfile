FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /agent-control-api ./cmd/api

FROM gcr.io/distroless/static-debian12
COPY --from=builder /agent-control-api /agent-control-api
EXPOSE 8080
ENTRYPOINT ["/agent-control-api"]
