# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .

# Build the Go app statically
RUN CGO_ENABLED=0 go build -o wap-tool .

# Final minimal container
FROM scratch
WORKDIR /app
COPY --from=builder /app/ttf /app/ttf
COPY --from=builder /app/wap-tool /app/wap-tool

ENTRYPOINT ["/app/wap-tool"]
