# Stage 1: Build Assets
FROM node:20.11-alpine AS asset-builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npx tailwindcss -i app/assets/app.css -o ./public/assets/styles.css
RUN npx esbuild app/assets/index.js --bundle --outdir=public/assets --minify

# Stage 2: Build Go Binary
FROM golang:1.26.2-alpine AS builder
RUN apk add --no-cache gcc musl-dev

# Install templ to generate view components
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1001

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Generate the Go code from .templ files
RUN templ generate

# Copy built assets from previous stage
COPY --from=asset-builder /app/public/assets ./public/assets
RUN GOOS=linux go build -o bin/app_prod ./cmd/app/main.go

# Stage 3: Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

# Copy binary and required runtime folders
COPY --from=builder /app/bin/app_prod ./app_prod
COPY --from=builder /app/public ./public

# Create an empty .env file so superkit's kit.Setup() doesn't crash
RUN touch .env

# Ensure config directory exists and copy the migration fallback
RUN mkdir -p /app/app/config
COPY --from=builder /app/app/config/config.json /app/app/config/config.json

# Ensure the upload directory exists for the volume mount
RUN mkdir -p /app/public/images

ENV DB_DRIVER=postgres
ENV DB_NAME=shop
ENV HTTP_LISTEN_ADDR=:3000

EXPOSE 3000
CMD ["./app_prod"]