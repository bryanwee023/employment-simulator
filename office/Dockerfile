FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o office .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/office .

EXPOSE 8080
ENTRYPOINT ["./office"]
