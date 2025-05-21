# current is all from Pelican

# Language
FROM golang:1.23-alpine

# Set working directory inside the container
WORKDIR /app

# Copy the entire repository
COPY . .

# Change to pelican directory
WORKDIR /app/owl

# Copy go mod files separately for better caching # commented out for now, not sure why
# RUN go mod download

# Build the Go application (static binary for minimal runtime dependencies)
RUN CGO_ENABLED=0 go build -o ./main

# Command to run the application
CMD ["./main"]
