// Command insuvit runs the Insuvit Teh Herbal branding website together
// with its admin panel and SQLite database.
//
// Configuration via environment variables (all optional):
//
//	ADDR            listen address              (default ":8080")
//	DB_PATH         SQLite file path            (default "insuvit.db")
//	WEB_DIR         templates + static dir      (default "web")
//	ADMIN_USER      seeded admin username       (default "admin")
//	ADMIN_PASS      seeded admin password       (default "insuvit123")
//	SESSION_SECRET  HMAC key for sessions       (default dev value)
package main

import (
	"context"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"insuvit/internal/auth"
	"insuvit/internal/db"
	"insuvit/internal/handlers"
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found or failed to load, falling back to OS environment variables")
	}

	// Dukungan untuk Render.com (Render menggunakan env PORT)
	port := env("PORT", "8080")
	addr := env("ADDR", ":"+port)
	
	dbPath := env("DB_PATH", "data/insuvit.db")
	webDir := env("WEB_DIR", "web")
	adminUser := env("ADMIN_USER", "admin")
	adminPass := env("ADMIN_PASS", "insuvit123")
	secret := env("SESSION_SECRET", "change-this-secret-in-production")

	// Database
	conn, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer conn.Close()
	if err := db.Migrate(conn); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := db.Seed(conn, adminUser, adminPass); err != nil {
		log.Fatalf("seed: %v", err)
	}
	store := db.NewStore(conn)

	// Templates
	publicTmpl, err := template.ParseGlob(filepath.Join(webDir, "templates", "public", "*.html"))
	if err != nil {
		log.Fatalf("parse public template: %v", err)
	}
	adminTmpl, err := template.ParseGlob(filepath.Join(webDir, "templates", "admin", "*.html"))
	if err != nil {
		log.Fatalf("parse admin templates: %v", err)
	}

	h := &handlers.Handlers{
		Store:    store,
		Sessions: auth.NewSessionManager(secret),
		Limiter:  auth.NewLoginLimiter(),
		Public:   publicTmpl,
		Admin:    adminTmpl,
		WebDir:   webDir,
	}

	mux := http.NewServeMux()

	// Static assets (css, js, images)
	staticFS := http.FileServer(http.Dir(filepath.Join(webDir, "static")))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticFS))

	// Public site + read-only API
	mux.HandleFunc("GET /{$}", h.Index)
	mux.HandleFunc("GET /api/products", h.APIProducts)
	mux.HandleFunc("GET /api/testimonials", h.APITestimonials)
	mux.HandleFunc("GET /api/faqs", h.APIFAQs)
	mux.HandleFunc("GET /api/settings", h.APISettings)

	// Redirect routes for analytics
	mux.HandleFunc("GET /r/shopee", h.RedirectShopee)
	mux.HandleFunc("GET /r/whatsapp", h.RedirectWhatsApp)

	// Admin auth
	mux.HandleFunc("GET /admin/login", h.LoginForm)
	mux.HandleFunc("POST /admin/login", h.LoginSubmit)
	mux.HandleFunc("POST /admin/logout", h.Logout)

	// Google OAuth
	mux.HandleFunc("GET /admin/auth/google", h.GoogleLoginStart)
	mux.HandleFunc("GET /admin/auth/google/callback", h.GoogleCallback)

	// Admin dashboard
	mux.HandleFunc("GET /admin", h.RequireAuth(h.Dashboard))
	mux.HandleFunc("GET /admin/{$}", h.RequireAuth(h.Dashboard))

	// Image upload (shared across forms)
	mux.HandleFunc("POST /admin/upload-image", h.RequireAuth(h.ImageUpload))

	// Admin products
	mux.HandleFunc("GET /admin/products", h.RequireAuth(h.ProductsList))
	mux.HandleFunc("GET /admin/products/new", h.RequireAuth(h.ProductNew))
	mux.HandleFunc("POST /admin/products", h.RequireAuth(h.ProductCreate))
	mux.HandleFunc("POST /admin/products/reorder", h.RequireAuth(h.ProductsReorder))
	mux.HandleFunc("GET /admin/products/{id}/edit", h.RequireAuth(h.ProductEdit))
	mux.HandleFunc("POST /admin/products/{id}", h.RequireAuth(h.ProductUpdate))
	mux.HandleFunc("POST /admin/products/{id}/delete", h.RequireAuth(h.ProductDelete))

	// Admin testimonials
	mux.HandleFunc("GET /admin/testimonials", h.RequireAuth(h.TestimonialsList))
	mux.HandleFunc("GET /admin/testimonials/new", h.RequireAuth(h.TestimonialNew))
	mux.HandleFunc("POST /admin/testimonials", h.RequireAuth(h.TestimonialCreate))
	mux.HandleFunc("POST /admin/testimonials/reorder", h.RequireAuth(h.TestimonialsReorder))
	mux.HandleFunc("GET /admin/testimonials/{id}/edit", h.RequireAuth(h.TestimonialEdit))
	mux.HandleFunc("POST /admin/testimonials/{id}", h.RequireAuth(h.TestimonialUpdate))
	mux.HandleFunc("POST /admin/testimonials/{id}/delete", h.RequireAuth(h.TestimonialDelete))

	// Admin FAQs
	mux.HandleFunc("GET /admin/faqs", h.RequireAuth(h.FAQsList))
	mux.HandleFunc("GET /admin/faqs/new", h.RequireAuth(h.FAQNew))
	mux.HandleFunc("POST /admin/faqs", h.RequireAuth(h.FAQCreate))
	mux.HandleFunc("POST /admin/faqs/reorder", h.RequireAuth(h.FAQsReorder))
	mux.HandleFunc("GET /admin/faqs/{id}/edit", h.RequireAuth(h.FAQEdit))
	mux.HandleFunc("POST /admin/faqs/{id}", h.RequireAuth(h.FAQUpdate))
	mux.HandleFunc("POST /admin/faqs/{id}/delete", h.RequireAuth(h.FAQDelete))

	// Admin settings
	mux.HandleFunc("GET /admin/settings", h.RequireAuth(h.SettingsForm))
	mux.HandleFunc("POST /admin/settings", h.RequireAuth(h.SettingsSave))
	mux.HandleFunc("POST /admin/settings/password", h.RequireAuth(h.SettingsPasswordSave))
	mux.HandleFunc("POST /admin/settings/avatar", h.RequireAuth(h.AvatarSave))

	// Wrap mux with RecordPageView middleware and then logRequests
	handler := h.RecordPageView(mux)

	srv := &http.Server{
		Addr:         addr,
		Handler:      logRequests(handler),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Create context that listens for the interrupt signal from the OS
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initializing the server in a goroutine so that it won't block the graceful shutdown handling below
	go func() {
		log.Printf("Insuvit server berjalan di http://localhost%s", addr)
		log.Printf("Halaman publik : http://localhost%s/", addr)
		log.Printf("Panel admin    : http://localhost%s/admin/login  (user: %s)", addr, adminUser)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Listen for the interrupt signal
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown
	stop()
	log.Println("Menutup server secara aman (Graceful Shutdown)...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	log.Println("Server berhasil dimatikan dengan aman.")
}

// logRequests is a tiny request logger middleware.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
