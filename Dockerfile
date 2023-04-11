# Use the official Golang image as a base image
FROM golang:1.20 as builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code to the working directory
COPY . .

# Build the Go app (use the correct path for your main.go)
RUN CGO_ENABLED=0 GOOS=linux go build -v -o main ./cmd/barcode_scanner/main.go

# Start a new stage from the scratch image
FROM scratch

# Copy the binary from the previous stage
COPY --from=builder /app/main /main

# Expose the port the app will run on
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/main"]