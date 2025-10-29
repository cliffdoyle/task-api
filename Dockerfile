# --- Stage 1: Build ---
# Use the official Golang Alpine image as the builder.
# Alpine Linux is much smaller than other distributions.
FROM golang:1.22-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed.
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app.
# CGO_ENABLED=0 creates a statically linked binary.
# -o task-api specifies the output file name.
# ./cmd/api is the path to our main package.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o task-api ./cmd/api


# --- Stage 2: Run ---
# Use a minimal Alpine image for the final stage.
# It does not contain any Go build tools, resulting in a much smaller image.
FROM alpine:latest

# Add ca-certificates for making HTTPS requests (good practice)
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /root/

# Copy the built binary from the 'builder' stage
COPY --from=builder /app/task-api .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
# This will be the container's entrypoint.
CMD ["./task-api"]