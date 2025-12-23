# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o autozap .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata bash

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/autozap .

# Copy example workflows
COPY --from=builder /app/workflows ./workflows

# Create directory for user workflows
RUN mkdir -p /workflows

# Expose any ports if needed (for future webhook support)
EXPOSE 8080

ENTRYPOINT ["./autozap"]
CMD ["--help"]
