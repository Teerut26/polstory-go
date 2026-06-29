FROM golang:1.23.4-alpine AS buildergo
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM node:22-alpine AS builderfrontend
WORKDIR /app
COPY /frontend/ .
RUN yarn global add pnpm && pnpm install --frozen-lockfile
RUN pnpm run build

FROM alpine:latest AS runner
RUN apk -U add exiftool
WORKDIR /app
COPY --from=buildergo /app/main .
COPY --from=buildergo /app/fonts ./fonts
COPY --from=builderfrontend /app/dist ./web

EXPOSE 3000

CMD ["./main"]