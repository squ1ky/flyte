FROM golang:1.25.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o user-service ./cmd/user

FROM alpine:3.19

WORKDIR /root

COPY --from=builder /app/user-service .
COPY --from=builder /app/migrations/user ./migrations/user

EXPOSE 50051

CMD ["./user-service"]