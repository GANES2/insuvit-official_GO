// Package handlers contains the HTTP handlers for the public site,
// the JSON API, and the admin panel.
package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"insuvit/internal/auth"
	"insuvit/internal/db"
	"insuvit/internal/models"
)

// Handlers holds shared dependencies for every route.
type Handlers struct {
	Store    *db.Store
	Sessions *auth.SessionManager
	Limiter  *auth.LoginLimiter
	Public   *template.Template // index page
	Admin    *template.Template // admin templates (parsed as a set)
	WebDir   string             // path to the web/ directory (for image uploads)
}

// Index renders the public branding page from database content.
func (h *Handlers) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	settings, err := h.Store.AllSettings()
	if err != nil {
		h.serverError(w, err)
		return
	}
	products, err := h.Store.ListProducts(true)
	if err != nil {
		h.serverError(w, err)
		return
	}
	testimonials, err := h.Store.ListTestimonials(true)
	if err != nil {
		h.serverError(w, err)
		return
	}
	faqs, err := h.Store.ListFAQs(true)
	if err != nil {
		h.serverError(w, err)
		return
	}
	data := models.PageData{
		Settings:     settings,
		Products:     products,
		Testimonials: testimonials,
		FAQs:         faqs,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Public.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("render index: %v", err)
	}
}

// ---------- JSON API (read-only, for future SPA / mobile use) ----------

func (h *Handlers) APIProducts(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.ListProducts(true)
	if err != nil {
		h.serverError(w, err)
		return
	}
	writeJSON(w, items)
}

func (h *Handlers) APITestimonials(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.ListTestimonials(true)
	if err != nil {
		h.serverError(w, err)
		return
	}
	writeJSON(w, items)
}

func (h *Handlers) APIFAQs(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.ListFAQs(true)
	if err != nil {
		h.serverError(w, err)
		return
	}
	writeJSON(w, items)
}

func (h *Handlers) APISettings(w http.ResponseWriter, r *http.Request) {
	s, err := h.Store.AllSettings()
	if err != nil {
		h.serverError(w, err)
		return
	}
	writeJSON(w, s)
}

// ---------- helpers ----------

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handlers) serverError(w http.ResponseWriter, err error) {
	log.Printf("server error: %v", err)
	http.Error(w, "Terjadi kesalahan pada server.", http.StatusInternalServerError)
}
