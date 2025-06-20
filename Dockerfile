# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go/go.mod go/go.sum ./
RUN go mod download
COPY go/ .

# Build the Go app statically
RUN CGO_ENABLED=0 go build -o wap-tool .

# Build frontend
FROM node:24-alpine as frontend

COPY ui/public/ /app/ui/public/
COPY ui/src/ /app/ui/src/
COPY ui/*.json ui/*.html ui/*.ts /app/ui/
COPY schema/wap.json /app/schema/wap.json
WORKDIR /app/ui/

RUN npm ci && npm run build-only

# Final minimal container
FROM scratch
WORKDIR /app
COPY --from=builder /app/ttf /app/ttf
COPY --from=builder /app/wap-tool /app/wap-tool

COPY --from=frontend /app/ui/dist/ /app/static

ENTRYPOINT ["/app/wap-tool", "web"]
