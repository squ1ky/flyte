FROM golang:1.25.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o gateway-service ./cmd/gateway

FROM alpine:3.19

WORKDIR /root

COPY --from=builder /app/gateway-service .

EXPOSE 8080

CMD ["./gateway-service"]