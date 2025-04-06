FROM golang:1.23-alpine

# Install ffmpeg if needed
RUN apk add --no-cache ffmpeg

# Set working directory for the entire project
WORKDIR /app

# Copy the entire repository
COPY . .

# Build the Go application from the pigeon directory 
RUN cd pigeon && CGO_ENABLED=0 go build -o ./main

# Set the working directory to where the binary is
WORKDIR /app/pigeon

# Command to run the application
CMD ["./main"]
