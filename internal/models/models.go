// Package models defines the data structures stored in the database
// and passed to the HTML templates.
package models

// Product is one item in the Shopee-linked catalog.
type Product struct {
	ID          int64
	Name        string
	Description string
	MetaLeft    string // left pill in the product-meta row, e.g. "Mulai Rp13.000*"
	MetaRight   string // right pill, e.g. "1 Pouch"
	Image       string // filename only, served from /static/images/
	ShopeeURL   string
	Badge       string // label shown on the image, e.g. "Tersedia di Shopee"
	Ribbon      string // optional corner ribbon, e.g. "Terlaris" ("" = hidden)
	SortOrder   int
	IsActive    bool
}

// Testimonial is a single customer quote on the public page.
type Testimonial struct {
	ID          int64
	Quote       string
	Author      string
	Role        string
	SortOrder   int
	IsPublished bool
}

// FAQ is one question/answer pair in the accordion.
type FAQ struct {
	ID        int64
	Question  string
	Answer    string
	SortOrder int
	IsActive  bool
}

// AdminUser is a login account for the admin panel.
type AdminUser struct {
	ID           int64
	Username     string
	PasswordHash string
}

// Settings is a flat key/value bag of editable site text
// (brand contact info, hero copy, P-IRT, etc.).
type Settings map[string]string

// PageData is the full payload handed to the public index template.
type PageData struct {
	Settings     Settings
	Products     []Product
	Testimonials []Testimonial
	FAQs         []FAQ
}
