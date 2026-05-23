# ── Stage 1: Build ─────────────────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

WORKDIR /app

# git needed for some go modules
RUN apk add --no-cache git

# Cache dependencies layer separately (faster rebuilds)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server main.go

# ── Stage 2: Runtime ────────────────────────────────────────────────────────
FROM alpine:3.19

WORKDIR /app

# ca-certificates needed for any HTTPS calls; tzdata for correct time zones
RUN apk add --no-cache ca-certificates tzdata

# Copy the compiled binary
COPY --from=builder /app/server .

# Create log storage dir
RUN mkdir -p storage/logs

EXPOSE 8080

CMD ["./server"]
