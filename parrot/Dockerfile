FROM golang:1.23-alpine

WORKDIR /app

# Copy the entire repository with proper structure
COPY . .

# This ensures that go.mod and related files are in the working directory
WORKDIR /app/parrot

# Now download dependencies
RUN go mod download

# Build the Go application
RUN CGO_ENABLED=0 go build -o ./main

# Command to run the application
CMD ["./main"]
