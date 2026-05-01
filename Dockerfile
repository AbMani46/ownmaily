# Build stage
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build Vue frontend
RUN apk add --no-cache nodejs npm
RUN npm install -g pnpm
RUN CI=true pnpm --dir frontend install
RUN pnpm --dir frontend build
# Build Go binary (embeds frontend/dist)
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/ownmaily ./cmd/server

# Runtime stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/bin/ownmaily .
EXPOSE 4400
CMD ["./ownmaily"]
