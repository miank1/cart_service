FROM golang:1.25.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o cart-service ./cmd/main.go

FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/cart-service .
EXPOSE 8083

CMD ["./cart-service"]
