# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/kafka-filter ./cmd/filter

# Runtime stage
FROM gcr.io/distroless/static-debian12
COPY --from=builder /out/kafka-filter /usr/local/bin/kafka-filter
COPY config/config.example.yaml /etc/kafka-bridge/config.yaml

ENTRYPOINT ["/usr/local/bin/kafka-filter"]
CMD ["-config", "/etc/kafka-bridge/config.yaml"]
