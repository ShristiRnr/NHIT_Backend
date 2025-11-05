FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o user-service ./services/user-service/cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/user-service .

EXPOSE 50051

CMD ["./user-service"]
