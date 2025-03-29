# ---- Step 1: Build Stage ----
FROM golang:1.23-alpine

# Install ffmpeg
RUN apk add --no-cache ffmpeg

# Set working directory inside the container
WORKDIR /app

# Copy the parent directory (muse) to /app
COPY ../. .

# Move to the pigeon directory
WORKDIR /app/pigeon

# Build the Go application 
RUN CGO_ENABLED=0 go build -o ./main

# Command to run the application
CMD ["./main"]
