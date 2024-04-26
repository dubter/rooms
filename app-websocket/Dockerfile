# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app-websocket
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /app-websocket/app /app-websocket/cmd/main.go

# Run stage
FROM alpine
WORKDIR /app-websocket
COPY --from=builder /app-websocket/app .

EXPOSE 443
ENTRYPOINT [ "./app", "-config", "/etc/app-websocket/dev.yaml", "-env", "/etc/app-websocket/.env"]