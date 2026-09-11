FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /task-manager ./cmd/server


FROM alpine:3.22

RUN addgroup -S appgroup && \
    adduser -S appuser -G appgroup && \
    apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /task-manager /app/task-manager

USER appuser:appgroup

EXPOSE 8080

ENTRYPOINT ["/app/task-manager"]