# Start from a Debian-based image with Go installed
FROM golang:latest AS build

# Set the working directory in the container
WORKDIR /app

# Copy the Go modules manifest and fetch dependencies
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Start a new stage from scratch
FROM alpine:latest  

# Set the working directory in the container
WORKDIR /root/

# Copy the pre-built binary from the previous stage
COPY --from=build /app/app .

# Expose port 3000 to the outside world
EXPOSE $APP_PORT

# Command to run the executable
CMD ["./app"]
