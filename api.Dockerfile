FROM golang:1.25.0-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

FROM alpine:latest
RUN adduser -D appuser
WORKDIR /app
COPY --from=builder /app/api .
COPY --from=builder /app/web ./web
COPY --chown=appuser:appuser configs ./configs
RUN mkdir -p storage && chown appuser:appuser storage
USER appuser
EXPOSE 8080
CMD ["./api"]