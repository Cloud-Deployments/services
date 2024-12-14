FROM golang:1.23-alpine AS builder

RUN mkdir -p /app
WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o ./bin/runner-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN mkdir -p /app
WORKDIR /app

COPY --from=builder /app/bin/runner-service /app/runner-service
COPY --from=builder /app/.env /app/.env

CMD ["./runner-service"]