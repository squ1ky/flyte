FROM golang:1.25.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o flight-service ./cmd/flight

FROM alpine:3.19

WORKDIR /root

COPY --from=builder /app/flight-service .
COPY --from=builder /app/migrations/flight ./migrations/flight

EXPOSE 50052

CMD ["./flight-service"]