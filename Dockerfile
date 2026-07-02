# Tahap 1: Build (Kompilasi kode Golang)
FROM golang:1.22-alpine AS builder

# Install gcc dan musl-dev karena SQLite membutuhkan CGO (C language bindings)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy file mod dan sum, lalu download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh kode sumber
COPY . .

# Kompilasi aplikasi dengan CGO_ENABLED=1 untuk SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -o insuvit-server cmd/insuvit/main.go

# Tahap 2: Production Image (Kecil dan ringan)
FROM alpine:latest

WORKDIR /app

# Copy binary dari tahap builder
COPY --from=builder /app/insuvit-server .

# Copy folder web (template HTML dan static assets)
COPY --from=builder /app/web ./web

# Buat folder data untuk SQLite dengan hak akses
RUN mkdir -p data && chmod 777 data

# Buka port 8080 (Bisa ditimpa oleh env PORT)
EXPOSE 8080

# Jalankan aplikasi
CMD ["./insuvit-server"]
