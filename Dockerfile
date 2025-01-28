FROM golang:1.23.4-alpine AS builderGO
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM node:20-alpine AS builderFrontend
WORKDIR /app
COPY /frontend/ .

FROM alpine:latest AS runner
RUN apk -U add exiftool
WORKDIR /app
COPY --from=builderGO /app/main .
COPY --from=builderGO /app/fonts ./fonts

EXPOSE 3000

# Command to run the executable
CMD ["./main"]