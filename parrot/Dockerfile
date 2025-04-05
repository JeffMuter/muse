# ---- Step 1: Build Stage ----
FROM golang:1.23-alpine

# Set working directory inside the container
WORKDIR /main

# Copy go mod files first
COPY go.mod go.sum ./

# Important: Copy the parent module that's being referenced via 'replace'
COPY .. /muse/

# Now modify the replace directive to point to the copied location
RUN go mod edit -replace=github.com/jeffmuter/muse=/muse

# Now download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 go build -o ./main

# Command to run the application
CMD ["./main"]
