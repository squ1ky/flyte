FROM golang:1.25.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o booking-service ./cmd/booking

FROM alpine:3.19

WORKDIR /root

COPY --from=builder /app/booking-service .
COPY --from=builder /app/migrations/booking ./migrations/booking

EXPOSE 50053

CMD ["./booking-service"]