# Start from the latest golang base image
FROM golang:1.23.4-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first; they are less frequently changed than source code, so Docker can cache this layer
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main .

FROM alpine:latest AS runner
RUN apk -U add exiftool

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/fonts /fonts

EXPOSE 3000

# Command to run the executable
CMD ["./main"]