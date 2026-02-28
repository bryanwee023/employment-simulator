FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o employee .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/employee .

EXPOSE 8080
ENTRYPOINT ["./employee"]
