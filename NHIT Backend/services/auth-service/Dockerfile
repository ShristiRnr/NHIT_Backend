FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy service and local module go.mod files
COPY services/auth-service/go.mod services/auth-service/go.sum ./services/auth-service/
COPY api/proto/go.mod api/proto/go.sum ./api/proto/

# Download dependencies for the auth-service module
RUN cd services/auth-service && go mod download

# Copy source code
COPY . .

# Build the service from within the module directory
RUN cd services/auth-service && CGO_ENABLED=0 GOOS=linux go build -o /app/auth-service ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/auth-service .

EXPOSE 50052

CMD ["./auth-service"]
