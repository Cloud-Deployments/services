FROM golang:1.23-alpine AS builder

RUN mkdir -p /app
WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o ./bin/coordinator-service


FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN mkdir -p /app
WORKDIR /app

COPY --from=builder /app/bin/coordinator-service /app/coordinator-service
COPY --from=builder /app/.env /app/.env

CMD ["./coordinator-service"]