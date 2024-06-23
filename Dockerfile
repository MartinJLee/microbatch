# Use the official Golang image as the base image
FROM golang:1.18-alpine AS builder

# Install gcc and other necessary build tools
RUN apk update && apk add --no-cache gcc musl-dev

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download the dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Run tests
RUN go test ./microbatch

# Build the Go application
RUN go build -o microbatch-service main.go

FROM golang:1.18-alpine AS runtime

# Set the working directory inside the container
WORKDIR /app

COPY --from=builder /app/microbatch-service .

# Expose the port the service will run on
EXPOSE 8080

# Set environment variables with default values
ENV BATCH_SIZE=5
ENV BATCH_INTERVAL=2

# Set the entry point to run the application
CMD ["./microbatch-service"]
