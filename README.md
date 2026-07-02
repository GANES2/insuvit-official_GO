# Insuvit Web Platform 🚀

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![HTMX](https://img.shields.io/badge/HTMX-1.9.10-336699?style=for-the-badge&logo=htmx)](https://htmx.org/)
[![Alpine.js](https://img.shields.io/badge/Alpine.js-3.13.3-8BC0D0?style=for-the-badge&logo=alpine.js)](https://alpinejs.dev/)
[![SQLite](https://img.shields.io/badge/SQLite-Database-003B57?style=for-the-badge&logo=sqlite)](https://www.sqlite.org/)

Platform profil perusahaan dan dashboard admin modern yang dibangun menggunakan arsitektur **Hypermedia-Driven Application (HDA)**. Insuvit memadukan kecepatan backend Golang murni dengan interaktivitas frontend yang ringan tanpa menggunakan *bloated SPA frameworks*.

## ✨ Fitur Utama (Features)

*   **🔒 Secure Admin Dashboard**: Autentikasi berbasis *Session* dengan proteksi enkripsi sandi **Bcrypt**.
*   **⚡ Hypermedia-Driven UI**: Navigasi *dashboard* tanpa ralat (*seamless*) menggunakan **HTMX** (tanpa perlu *reload* halaman penuh).
*   **🐻 Easter Egg UI**: Desain menu aktif *Sidebar* bergaya "Beruang Melet" yang *seamless* dengan gradasi halaman.
*   **📦 Dynamic Database**: Manajemen FAQ, Testimoni (dengan fitur geser/Reorder menggunakan SortableJS), dan Pengaturan Situs menggunakan **SQLite3**.
*   **📊 Analytics Tracking**: Perekaman otomatis (*click tracking*) setiap interaksi pengguna ke tombol WhatsApp atau Shopee.
*   **🖼️ Image Uploads**: Sistem *Multipart Form Data* untuk mengatur Avatar Admin dan Testimoni secara aman.

## 🏗️ Struktur Proyek (Architecture)

Proyek ini menggunakan *Standard Go Project Layout*:
```text
.
├── cmd/
│   └── insuvit/          # Titik utama (Entry point) aplikasi
├── internal/
│   ├── db/               # Logika database, schema, dan eksekusi SQL
│   ├── handlers/         # Controller untuk rute Publik & Admin
│   ├── middleware/       # Autentikasi dan proteksi keamanan HTTP
│   └── models/           # Definisi struktur data (Struct)
├── web/
│   ├── static/           # Aset statis (CSS murni, JS vanilla, images)
│   └── templates/        # Template HTML (Go html/template)
└── Makefile              # Kumpulan perintah otomatisasi
```

## 🛠️ Cara Menjalankan (Getting Started)

Proyek ini dilengkapi dengan `Makefile` untuk mempermudah eksekusi layaknya *developer* profesional.

1. **Persiapan:**
   Pastikan Anda sudah menginstal [Go](https://go.dev/doc/install) dan `gcc` (untuk SQLite CGO).

2. **Menjalankan Server (Mode Development):**
   ```bash
   make dev
   ```
   *Server akan berjalan di `http://localhost:8080`.*

3. **Membangun File Binary (Mode Production):**
   ```bash
   make build
   ```
   *Perintah ini akan mengompilasi kode Anda menjadi satu file executable bernama `insuvit-server` yang siap di-deploy ke Cloud.*

4. **Menampilkan Bantuan Makefile:**
   ```bash
   make help
   ```

## 👨‍💻 Kontributor

Dikembangkan dengan ❤️ dan dedikasi penuh terhadap pengalaman pengguna (UI/UX) dan *Clean Code*.
