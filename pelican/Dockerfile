FROM golang:1.23-alpine

# Install ffmpeg
RUN apk add --no-cache ffmpeg

# Set working directory inside the container
WORKDIR /app

# Copy the entire repository
COPY . .

# Change to pelican directory
WORKDIR /app/pelican

# Copy go mod files separately for better caching
# RUN go mod download

# Build the Go application (static binary for minimal runtime dependencies)
RUN CGO_ENABLED=0 go build -o ./main

# Command to run the application
CMD ["./main"]
