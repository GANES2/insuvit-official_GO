// Package db handles the SQLite connection, schema migration, and the
// initial seed data (taken from the current Insuvit branding site).
package db

import (
	"database/sql"
	"fmt"
	"log"

	"insuvit/internal/auth"
	"insuvit/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

// Open returns a configured SQLite connection at the given file path.
func Open(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_busy_timeout=5000&_foreign_keys=on", path)
	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1) // SQLite is single-writer; keep it simple
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	return conn, nil
}

const schema = `
CREATE TABLE IF NOT EXISTS admin_users (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  username      TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS settings (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS products (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  name        TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  meta_left   TEXT NOT NULL DEFAULT '',
  meta_right  TEXT NOT NULL DEFAULT '',
  image       TEXT NOT NULL DEFAULT '',
  shopee_url  TEXT NOT NULL DEFAULT '',
  badge       TEXT NOT NULL DEFAULT '',
  ribbon      TEXT NOT NULL DEFAULT '',
  sort_order  INTEGER NOT NULL DEFAULT 0,
  is_active   INTEGER NOT NULL DEFAULT 1,
  created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS testimonials (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  quote        TEXT NOT NULL,
  author       TEXT NOT NULL DEFAULT '',
  role         TEXT NOT NULL DEFAULT '',
  sort_order   INTEGER NOT NULL DEFAULT 0,
  is_published INTEGER NOT NULL DEFAULT 1,
  created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS faqs (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  question   TEXT NOT NULL,
  answer     TEXT NOT NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  is_active  INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS page_views (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  path       TEXT NOT NULL,
  user_agent TEXT NOT NULL,
  ip_address TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS click_events (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  target     TEXT NOT NULL,
  url        TEXT NOT NULL,
  user_agent TEXT NOT NULL,
  ip_address TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

// Migrate creates all tables if they do not exist.
func Migrate(conn *sql.DB) error {
	_, err := conn.Exec(schema)
	return err
}

// Seed fills the database with the initial Insuvit content on first run.
// It is idempotent: it only inserts into a table when that table is empty.
func Seed(conn *sql.DB, adminUser, adminPass string) error {
	if err := seedAdmin(conn, adminUser, adminPass); err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}
	if err := seedSettings(conn); err != nil {
		return fmt.Errorf("seed settings: %w", err)
	}
	if err := seedProducts(conn); err != nil {
		return fmt.Errorf("seed products: %w", err)
	}
	if err := seedTestimonials(conn); err != nil {
		return fmt.Errorf("seed testimonials: %w", err)
	}
	if err := seedFAQs(conn); err != nil {
		return fmt.Errorf("seed faqs: %w", err)
	}
	return nil
}

func isEmpty(conn *sql.DB, table string) (bool, error) {
	var n int
	if err := conn.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		return false, err
	}
	return n == 0, nil
}

func seedAdmin(conn *sql.DB, user, pass string) error {
	empty, err := isEmpty(conn, "admin_users")
	if err != nil || !empty {
		return err
	}
	hash, err := auth.HashPassword(pass)
	if err != nil {
		return err
	}
	_, err = conn.Exec("INSERT INTO admin_users (username, password_hash) VALUES (?, ?)", user, hash)
	if err == nil {
		log.Printf("seeded admin user %q", user)
	}
	return err
}

func seedSettings(conn *sql.DB) error {
	empty, err := isEmpty(conn, "settings")
	if err != nil || !empty {
		return err
	}
	defaults := map[string]string{
		"site_title":      "Insuvit Teh Herbal - Website Branding UMKM",
		"hero_subtitle":   "Sehat Tanpa Obat Kimia",
		"hero_desc":       "UMKM teh herbal berbahan tumbuhan alami yang hadir untuk mendukung gaya hidup sehat, terutama bagi masyarakat yang peduli pada keseimbangan metabolisme dan kadar gula darah. Website ini menjadi pusat informasi resmi; pembelian tetap diarahkan ke toko Insuvit di Shopee.",
		"footer_desc":     "Usaha mikro yang memproduksi minuman teh herbal berbahan dasar tumbuhan alami, diformulasikan khusus untuk mendukung gaya hidup sehat dan keseimbangan metabolisme tubuh.",
		"shopee_url":      "https://shopee.co.id/insuvit_official",
		"whatsapp_link":   "https://wa.me/6281285340838",
		"whatsapp_number": "+62 812-8534-0838",
		"instagram_url":   "https://instagram.com/insuvit_official",
		"pirt":            "5083202011298-29",
		"rating":          "4.9/5",
		"followers":       "3,7RB",
	}
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	for k, v := range defaults {
		if _, err := tx.Exec("INSERT INTO settings (key, value) VALUES (?, ?)", k, v); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func seedProducts(conn *sql.DB) error {
	empty, err := isEmpty(conn, "products")
	if err != nil || !empty {
		return err
	}
	shopee := "https://shopee.co.id/insuvit_official"
	items := []models.Product{
		{Name: "Teh Insuvit Herbal — Isi 15 Teabag", Description: "Kemasan pouch isi 15 teh celup herbal rempah, cocok untuk mencoba pertama kali sebelum konsumsi rutin.", MetaLeft: "Mulai Rp13.000*", MetaRight: "1 Pouch", Image: "product-1.png", ShopeeURL: shopee, Badge: "Tersedia di Shopee", Ribbon: "Terlaris", SortOrder: 1, IsActive: true},
		{Name: "Teh Insuvit Herbal Original — Isi 30 Teabag", Description: "Kemasan satu bungkus isi 30 teh celup, varian original untuk konsumsi rutin sesuai petunjuk pada kemasan.", MetaLeft: "1 Bungkus", MetaRight: "Tersedia di Shopee", Image: "product-2.png", ShopeeURL: shopee, Badge: "Tersedia di Shopee", Ribbon: "", SortOrder: 2, IsActive: true},
		{Name: "Paket Isi Banyak — hingga 160 Teabag", Description: "Pilihan isi besar untuk persediaan konsumsi rutin, dengan harga mengikuti toko resmi Shopee.", MetaLeft: "Stok rutin", MetaRight: "Cek harga di Shopee", Image: "product-3.png", ShopeeURL: shopee, Badge: "Tersedia di Shopee", Ribbon: "", SortOrder: 3, IsActive: true},
		{Name: "Paket Hemat Insuvit", Description: "Bundling teh celup herbal untuk persediaan konsumsi rutin pagi dan malam sesuai aturan pakai pada kemasan, dengan harga mengikuti toko resmi Shopee.", MetaLeft: "Bundling", MetaRight: "COD Shopee", Image: "product-4.png", ShopeeURL: shopee, Badge: "Tersedia di Shopee", Ribbon: "", SortOrder: 4, IsActive: true},
	}
	for _, p := range items {
		if _, err := conn.Exec(
			`INSERT INTO products (name, description, meta_left, meta_right, image, shopee_url, badge, ribbon, sort_order, is_active)
			 VALUES (?,?,?,?,?,?,?,?,?,?)`,
			p.Name, p.Description, p.MetaLeft, p.MetaRight, p.Image, p.ShopeeURL, p.Badge, p.Ribbon, p.SortOrder, p.IsActive,
		); err != nil {
			return err
		}
	}
	return nil
}

func seedTestimonials(conn *sql.DB) error {
	empty, err := isEmpty(conn, "testimonials")
	if err != nil || !empty {
		return err
	}
	items := []models.Testimonial{
		{Quote: "Praktis tinggal seduh, jadi tidak perlu repot menyiapkan bahan herbal sendiri.", Author: "Ulasan Konsumen", Role: "Pembeli di Marketplace", SortOrder: 1, IsPublished: true},
		{Quote: "Pembelian terasa lebih aman karena diarahkan ke toko resmi di marketplace.", Author: "Ulasan Konsumen", Role: "Pembeli di Marketplace", SortOrder: 2, IsPublished: true},
		{Quote: "Informasi produk, cara beli, dan kontak resmi jadi lebih mudah ditemukan.", Author: "Ulasan Konsumen", Role: "Pembeli di Marketplace", SortOrder: 3, IsPublished: true},
	}
	for _, t := range items {
		if _, err := conn.Exec(
			`INSERT INTO testimonials (quote, author, role, sort_order, is_published) VALUES (?,?,?,?,?)`,
			t.Quote, t.Author, t.Role, t.SortOrder, t.IsPublished,
		); err != nil {
			return err
		}
	}
	return nil
}

func seedFAQs(conn *sql.DB) error {
	empty, err := isEmpty(conn, "faqs")
	if err != nil || !empty {
		return err
	}
	items := []models.FAQ{
		{Question: "Apa itu Insuvit Teh Herbal?", Answer: "Insuvit adalah minuman teh herbal dari tumbuhan alami yang diformulasikan untuk mendukung gaya hidup sehat dan membantu menjaga keseimbangan metabolisme tubuh.", SortOrder: 1, IsActive: true},
		{Question: "Siapa saja yang boleh mengonsumsi Insuvit?", Answer: "Target utama Insuvit adalah masyarakat umum di seluruh Indonesia yang peduli kesehatan dan ingin menjaga kebugaran serta kadar gula darah tetap stabil secara alami sebagai bagian dari gaya hidup sehat.", SortOrder: 2, IsActive: true},
		{Question: "Apakah produk Insuvit aman dan legal?", Answer: "Informasi P-IRT yang digunakan adalah 5083202011298-29 berdasarkan data produk di Shopee. Informasi halal dapat ditambahkan setelah dokumen resmi owner tersedia.", SortOrder: 3, IsActive: true},
		{Question: "Apakah saya bisa membeli langsung di website ini?", Answer: "Tidak. Website ini berfungsi sebagai media informasi dan branding. Untuk keamanan serta kenyamanan transaksi, pembelian diarahkan ke toko resmi Insuvit di marketplace, terutama Shopee.", SortOrder: 4, IsActive: true},
		{Question: "Di mana saja saya bisa membeli produk Insuvit?", Answer: "Insuvit aktif menggunakan Shopee sebagai kanal utama. Konsumen juga dapat menghubungi WhatsApp admin sebagai kontak pendukung.", SortOrder: 5, IsActive: true},
		{Question: "Apakah website ini sudah memakai pixel atau analytics?", Answer: "Pixel dan analytics belum diaktifkan. Integrasi Meta Pixel, Shopee Pixel, atau TikTok Pixel dapat dikembangkan pada tahap berikutnya dengan memperhatikan izin pengguna dan kebijakan privasi.", SortOrder: 6, IsActive: true},
		{Question: "Apakah teh herbal ini bisa menggantikan obat medis?", Answer: "Tidak. Insuvit Teh Herbal bukan pengganti obat atau konsultasi medis. Produk ini diposisikan sebagai minuman herbal untuk mendukung gaya hidup sehat. Konsumen dengan kondisi kesehatan tertentu disarankan berkonsultasi dengan tenaga kesehatan.", SortOrder: 7, IsActive: true},
	}
	for _, f := range items {
		if _, err := conn.Exec(
			`INSERT INTO faqs (question, answer, sort_order, is_active) VALUES (?,?,?,?)`,
			f.Question, f.Answer, f.SortOrder, f.IsActive,
		); err != nil {
			return err
		}
	}
	return nil
}
