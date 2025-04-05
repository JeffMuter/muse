FROM golang:1.23-alpine

# Install ffmpeg if needed
RUN apk add --no-cache ffmpeg

# Set working directory
WORKDIR /app

# Copy the entire repository
COPY . .

# Move to the pigeon directory
WORKDIR /app/pigeon

# Build the Go application 
RUN CGO_ENABLED=0 go build -o ./main

# Command to run the application
CMD ["./main"]
