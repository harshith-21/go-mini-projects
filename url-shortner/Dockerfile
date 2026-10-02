FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY main.go .
COPY templates ./templates

RUN go build -o url-shortener main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/url-shortener .
COPY --from=builder /app/templates ./templates

EXPOSE 8080

CMD ["./url-shortener"]