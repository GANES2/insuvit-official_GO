.PHONY: all build run dev clean

# Nama output binary
BINARY_NAME=insuvit-server

all: build

## build: Kompilasi aplikasi menjadi file binary siap pakai
build:
	@echo "Membangun aplikasi..."
	go build -o $(BINARY_NAME) cmd/insuvit/main.go
	@echo "Aplikasi berhasil dibangun: $(BINARY_NAME)"

## run: Menjalankan aplikasi dari file binary (production mode)
run: build
	@echo "Menjalankan aplikasi..."
	./$(BINARY_NAME)

## dev: Menjalankan aplikasi untuk pengembangan (development mode)
dev:
	@echo "Memulai mode pengembangan (Hot-reload manual)..."
	go run cmd/insuvit/main.go

## clean: Membersihkan file binary dan cache
clean:
	@echo "Membersihkan sistem..."
	go clean
	rm -f $(BINARY_NAME)
	@echo "Selesai."

## help: Menampilkan bantuan perintah yang tersedia
help:
	@echo "Perintah Makefile yang tersedia:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
