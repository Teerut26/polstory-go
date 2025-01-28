FROM golang:1.23.4-alpine AS builderGO
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM node:20-alpine AS builderFrontend
WORKDIR /app
COPY /frontend/ .
RUN yarn global add pnpm && pnpm install --frozen-lockfile
RUN pnpm run build

FROM alpine:latest AS runner
RUN apk -U add exiftool
WORKDIR /app
COPY --from=builderGO /app/main .
COPY --from=builderGO /app/fonts ./fonts
COPY --from=builderFrontend /app/dist ./web

EXPOSE 3000

CMD ["./main"]