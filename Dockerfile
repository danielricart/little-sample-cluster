FROM golang:1.25.4-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum* ./

# Download all dependencies
RUN go mod tidy && go mod download

# Copy the source code
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -ldflags="-s -w" -o little-sample-cluster .

# Start a new stage with minimal image
FROM scratch
LABEL org.opencontainers.image.source=https://github.com/danielricart/little-sample-cluster
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/little-sample-cluster .

# Expose port 8080
EXPOSE 8089

# Environment variables with defaults
ENV DB_HOST="localhost"
ENV DB_USERNAME="root"
ENV DB_PASSWORD=""
ENV DB_NAME=""
ENV DB_PORT=3306
ENV SERVER_PORT=8089

# Run the binary
CMD ["./little-sample-cluster"]
