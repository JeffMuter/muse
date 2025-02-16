FROM golang:1.23
WORKDIR /build
COPY ./ ./
RUN go mod download
RUN CGO_ENABLED=0 go build -o ./main
