package handlers

import (
	"log"
	"net/http"
	"strings"
)

// RecordPageView middleware records a visit to the given path.
func (h *Handlers) RecordPageView(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only record GET requests to public HTML pages (not static assets or admin)
		if r.Method == http.MethodGet && !strings.HasPrefix(r.URL.Path, "/static/") && !strings.HasPrefix(r.URL.Path, "/admin") {
			ip := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ip = forwarded
			}
			err := h.Store.RecordPageView(r.URL.Path, r.UserAgent(), ip)
			if err != nil {
				log.Printf("failed to record page view: %v", err)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RedirectShopee records a click event for Shopee and redirects the user.
func (h *Handlers) RedirectShopee(w http.ResponseWriter, r *http.Request) {
	settings, err := h.Store.AllSettings()
	if err != nil {
		h.serverError(w, err)
		return
	}

	url := settings["shopee_url"]

	// Check if specific product shopee URL is provided
	productID := r.URL.Query().Get("id")
	if productID != "" {
		// Attempt to get product URL
		// We could fetch product by ID here, but for simplicity we rely on the default store URL
		// if the product URL is empty.
	}

	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}

	err = h.Store.RecordClickEvent("shopee", url, r.UserAgent(), ip)
	if err != nil {
		log.Printf("failed to record shopee click: %v", err)
	}

	http.Redirect(w, r, url, http.StatusFound)
}

// RedirectWhatsApp records a click event for WhatsApp and redirects the user.
func (h *Handlers) RedirectWhatsApp(w http.ResponseWriter, r *http.Request) {
	settings, err := h.Store.AllSettings()
	if err != nil {
		h.serverError(w, err)
		return
	}

	url := settings["whatsapp_link"]

	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}

	err = h.Store.RecordClickEvent("whatsapp", url, r.UserAgent(), ip)
	if err != nil {
		log.Printf("failed to record whatsapp click: %v", err)
	}

	http.Redirect(w, r, url, http.StatusFound)
}
