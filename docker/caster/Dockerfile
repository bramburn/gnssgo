FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /ntrip-caster ./cmd/ntrip-caster

# Create a minimal image
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /ntrip-caster .

# Expose the default NTRIP port
EXPOSE 2101

# Run the application
CMD ["./ntrip-caster"]
