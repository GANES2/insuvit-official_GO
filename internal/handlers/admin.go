package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"insuvit/internal/auth"
	"insuvit/internal/models"
)

// adminView is the data passed to admin templates.
type adminView struct {
	Title        string
	User         string
	Section      string
	Error        string
	Success      string
	IsNew        bool
	CSRFToken    string
	Products     []models.Product
	Product      models.Product
	Testimonials []models.Testimonial
	Testimonial  models.Testimonial
	FAQs         []models.FAQ
	FAQ          models.FAQ
	Settings     models.Settings
	Analytics    map[string]int
}

func (h *Handlers) renderAdmin(w http.ResponseWriter, r *http.Request, name string, v adminView) {
	v.CSRFToken = h.Sessions.CSRFToken(r)
	if v.Settings == nil {
		v.Settings, _ = h.Store.AllSettings()
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Admin.ExecuteTemplate(w, name, v); err != nil {
		log.Printf("render admin %q: %v", name, err)
	}
}

// RequireAuth wraps a handler, redirecting to the login page if there is no
// valid session. It also validates CSRF tokens on all non-multipart POST requests.
func (h *Handlers) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := h.Sessions.Validate(r); err != nil {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		if r.Method == http.MethodPost && !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/") {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
			if !h.Sessions.ValidateCSRF(r) {
				http.Error(w, "Permintaan tidak valid (CSRF). Silakan refresh halaman dan coba lagi.", http.StatusForbidden)
				return
			}
		}
		next(w, r)
	}
}

func (h *Handlers) currentUser(r *http.Request) string {
	u, _ := h.Sessions.Validate(r)
	return u
}

func clientIP(r *http.Request) string {
	if addr := r.RemoteAddr; addr != "" {
		if i := strings.LastIndex(addr, ":"); i >= 0 {
			return addr[:i]
		}
		return addr
	}
	return "unknown"
}

// ---------- Auth ----------

func (h *Handlers) LoginForm(w http.ResponseWriter, r *http.Request) {
	if _, err := h.Sessions.Validate(r); err == nil {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	settings, _ := h.Store.AllSettings()
	h.renderAdmin(w, r, "login", adminView{Title: "Masuk Admin", Settings: settings})
}

func (h *Handlers) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)

	if !h.Limiter.Allow(ip) {
		remaining := h.Limiter.RemainingBlock(ip)
		mins := int(math.Ceil(remaining.Minutes()))
		msg := fmt.Sprintf("Terlalu banyak percobaan gagal. Coba lagi dalam %d menit.", mins)
		h.renderAdmin(w, r, "login", adminView{Title: "Masuk Admin", Error: msg})
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	user, err := h.Store.GetAdminByUsername(username)
	if err != nil || !auth.CheckPassword(password, user.PasswordHash) {
		h.Limiter.RecordFailure(ip)
		h.renderAdmin(w, r, "login", adminView{Title: "Masuk Admin", Error: "Username atau kata sandi salah."})
		return
	}
	h.Limiter.RecordSuccess(ip)
	h.Sessions.Create(w, user.Username)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	h.Sessions.Destroy(w)
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

// ---------- Dashboard ----------

func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	products, _ := h.Store.ListProducts(false)
	testimonials, _ := h.Store.ListTestimonials(false)
	faqs, _ := h.Store.ListFAQs(false)
	stats, _ := h.Store.GetAnalyticsStats()
	settings, _ := h.Store.AllSettings()
	h.renderAdmin(w, r, "dashboard", adminView{
		Title:        "Dashboard",
		User:         h.currentUser(r),
		Section:      "dashboard",
		Products:     products,
		Testimonials: testimonials,
		FAQs:         faqs,
		Analytics:    stats,
		Settings:     settings,
	})
}

// ---------- Products ----------

func (h *Handlers) ProductsList(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.ListProducts(false)
	if err != nil {
		h.serverError(w, err)
		return
	}
	h.renderAdmin(w, r, "products", adminView{Title: "Produk", User: h.currentUser(r), Section: "products", Products: items})
}

func (h *Handlers) ProductNew(w http.ResponseWriter, r *http.Request) {
	h.renderAdmin(w, r, "product_form", adminView{
		Title:   "Tambah Produk",
		User:    h.currentUser(r),
		Section: "products",
		IsNew:   true,
		Product: models.Product{IsActive: true, Badge: "Tersedia di Shopee"},
	})
}

func (h *Handlers) ProductCreate(w http.ResponseWriter, r *http.Request) {
	p := productFromForm(r)
	if _, err := h.Store.CreateProduct(p); err != nil {
		h.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/admin/products", http.StatusSeeOther)
}

func (h *Handlers) ProductEdit(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	p, err := h.Store.GetProduct(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	h.renderAdmin(w, r, "product_form", adminView{Title: "Ubah Produk", User: h.currentUser(r), Section: "products", Product: p})
}

func (h *Handlers) ProductUpdate(w http.ResponseWriter, r *http.Request) {
	p := productFromForm(r)
	p.ID = pathID(r)
	if err := h.Store.UpdateProduct(p); err != nil {
		h.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/admin/products", http.StatusSeeOther)
}

func (h *Handlers) ProductDelete(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.DeleteProduct(pathID(r)); err != nil {
		h.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/admin/products", http.StatusSeeOther)
}

func (h *Handlers) ProductsReorder(w http.ResponseWriter, r *http.Request) {
	ids := parseIDList(r.FormValue("ids"))
	w.Header().Set("Content-Type", "application/json")
	if len(ids) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"ok":false}`))
		return
	}
	if err := h.Store.ReorderProducts(ids); err != nil {
		log.Printf("reorder products: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"ok":false}`))
		return
	}
	w.Write([]byte(`{"ok":true}`))
}

func productFromForm(r *http.Request) models.Product {
	return models.Product{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		MetaLeft:    r.FormValue("meta_left"),
		MetaRight:   r.FormValue("meta_right"),
		Image:       r.FormValue("image"),
		ShopeeURL:   r.FormValue("shopee_url"),
		Badge:       r.FormValue("badge"),
		Ribbon:      r.FormValue("ribbon"),
		SortOrder:   formInt(r, "sort_order"),
		IsActive:    formBool(r, "is_active"),
	}
}

// ---------- Testimonials ----------

func (h *Handlers) TestimonialsList(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.ListTestimonials(false)
	if err != nil {
		h.serverError(w, err)
		return
	}
	h.renderAdmin(w, r, "testimonials", adminView{Title: "Testimoni", User: h.currentUser(r), Section: "testimonials", Testimonials: items})
}

func (h *Handlers) TestimonialNew(w http.ResponseWriter, r *http.Request) {
	h.renderAdmin(w, r, "testimonial_form", adminView{
		Title:       "Tambah Testimoni",
		User:        h.currentUser(r),
		Section:     "testimonials",
		IsNew:       true,
		Testimonial: models.Testimonial{IsPublished: true, Author: "Ulasan Konsumen", Role: "Pembeli di Marketplace"},
	})
}

func (h *Handlers) TestimonialCreate(w http.ResponseWriter, r *http.Request) {
	t := testimonialFromForm(r)
	if _, err := h.Store.CreateTestimonial(t); err != nil {
		h.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/admin/testimonials", http.StatusSeeOther)
}

func (h *Handlers) TestimonialEdit(w http.ResponseWriter, r *http.Request) {
	t, err := h.Store.GetTestimonial(pathID(r))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	h.renderAdmin(w, r, "testimonial_form", adminView{Title: "Ubah Testimoni", User: h.currentUser(r), Section: "testimonials", Testimonial: t})
}

func (h *Handlers) TestimonialUpdate(w http.ResponseWriter, r *http.Request) {
	t := testimonialFromForm(r)
	t.ID = pathID(r)
	if err := h.Store.UpdateTestimonial(t); err != nil {
		h.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/admin/testimonials", http.StatusSeeOther)
}

func (h *Handlers) TestimonialDelete(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.DeleteTestimonial(pathID(r)); err != nil {
		h.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/admin/testimonials", http.StatusSeeOther)
}

func (h *Handlers) TestimonialsReorder(w http.ResponseWriter, r *http.Request) {
	ids := parseIDList(r.FormValue("ids"))
	w.Header().Set("Content-Type", "application/json")
	if len(ids) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"ok":false}`))
		return
	}
	if err := h.Store.ReorderTestimonials(ids); err != nil {
		log.Printf("reorder testimonials: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"ok":false}`))
		return
	}
	w.Write([]byte(`{"ok":true}`))
}

func testimonialFromForm(r *http.Request) models.Testimonial {
	return models.Testimonial{
		Quote:       r.FormValue("quote"),
		Author:      r.FormValue("author"),
		Role:        r.FormValue("role"),
		SortOrder:   formInt(r, "sort_order"),
		IsPublished: formBool(r, "is_published"),
	}
}

// ---------- FAQs ----------

func (h *Handlers) FAQsList(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.ListFAQs(false)
	if err != nil {
		h.serverError(w, err)
		return
	}
	h.renderAdmin(w, r, "faqs", adminView{Title: "FAQ", User: h.currentUser(r), Section: "faqs", FAQs: items})
}

func (h *Handlers) FAQNew(w http.ResponseWriter, r *http.Request) {
	h.renderAdmin(w, r, "faq_form", adminView{
		Title:   "Tambah FAQ",
		User:    h.currentUser(r),
		Section: "faqs",
		IsNew:   true,
		FAQ:     models.FAQ{IsActive: true},
	})
}

func (h *Handlers) FAQCreate(w http.ResponseWriter, r *http.Request) {
	f := faqFromForm(r)
	if _, err := h.Store.CreateFAQ(f); err != nil {
		h.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/admin/faqs", http.StatusSeeOther)
}

func (h *Handlers) FAQEdit(w http.ResponseWriter, r *http.Request) {
	f, err := h.Store.GetFAQ(pathID(r))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	h.renderAdmin(w, r, "faq_form", adminView{Title: "Ubah FAQ", User: h.currentUser(r), Section: "faqs", FAQ: f})
}

func (h *Handlers) FAQUpdate(w http.ResponseWriter, r *http.Request) {
	f := faqFromForm(r)
	f.ID = pathID(r)
	if err := h.Store.UpdateFAQ(f); err != nil {
		h.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/admin/faqs", http.StatusSeeOther)
}

func (h *Handlers) FAQDelete(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.DeleteFAQ(pathID(r)); err != nil {
		h.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/admin/faqs", http.StatusSeeOther)
}

func (h *Handlers) FAQsReorder(w http.ResponseWriter, r *http.Request) {
	ids := parseIDList(r.FormValue("ids"))
	w.Header().Set("Content-Type", "application/json")
	if len(ids) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"ok":false}`))
		return
	}
	if err := h.Store.ReorderFAQs(ids); err != nil {
		log.Printf("reorder faqs: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"ok":false}`))
		return
	}
	w.Write([]byte(`{"ok":true}`))
}

func faqFromForm(r *http.Request) models.FAQ {
	return models.FAQ{
		Question:  r.FormValue("question"),
		Answer:    r.FormValue("answer"),
		SortOrder: formInt(r, "sort_order"),
		IsActive:  formBool(r, "is_active"),
	}
}

// ---------- Settings ----------

func (h *Handlers) SettingsForm(w http.ResponseWriter, r *http.Request) {
	s, err := h.Store.AllSettings()
	if err != nil {
		h.serverError(w, err)
		return
	}
	v := adminView{Title: "Pengaturan", User: h.currentUser(r), Section: "settings", Settings: s}
	if r.URL.Query().Get("success") == "1" {
		v.Success = "Pengaturan berhasil disimpan."
	}
	h.renderAdmin(w, r, "settings", v)
}

func (h *Handlers) SettingsSave(w http.ResponseWriter, r *http.Request) {
	keys := []string{
		"site_title", "meta_description", "meta_keywords",
		"hero_title", "hero_subtitle", "hero_desc", "footer_desc",
		"shopee_url", "whatsapp_link", "whatsapp_number", "instagram_url",
		"tiktok_url", "tokopedia_url",
		"pirt", "rating", "followers",
	}
	values := map[string]string{}
	for _, k := range keys {
		values[k] = r.FormValue(k)
	}
	if err := h.Store.UpdateSettings(values); err != nil {
		h.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/admin/settings?success=1", http.StatusSeeOther)
}

func (h *Handlers) AvatarSave(w http.ResponseWriter, r *http.Request) {
	filename := r.FormValue("admin_photo")
	if err := h.Store.UpdateSettings(map[string]string{"admin_photo": filename}); err != nil {
		h.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/admin/settings?success=1", http.StatusSeeOther)
}

func (h *Handlers) SettingsPasswordSave(w http.ResponseWriter, r *http.Request) {
	oldPass := r.FormValue("old_password")
	newPass := r.FormValue("new_password")
	confirmPass := r.FormValue("confirm_password")

	s, _ := h.Store.AllSettings()
	username := h.currentUser(r)

	renderErr := func(msg string) {
		h.renderAdmin(w, r, "settings", adminView{
			Title: "Pengaturan", User: username, Section: "settings", Settings: s,
			Error: msg,
		})
	}

	if len(newPass) < 8 {
		renderErr("Kata sandi baru minimal 8 karakter.")
		return
	}
	if !hasLetter(newPass) || !hasDigit(newPass) {
		renderErr("Kata sandi baru harus mengandung minimal satu huruf dan satu angka.")
		return
	}
	if newPass != confirmPass {
		renderErr("Konfirmasi kata sandi tidak cocok.")
		return
	}

	user, err := h.Store.GetAdminByUsername(username)
	if err != nil || !auth.CheckPassword(oldPass, user.PasswordHash) {
		renderErr("Kata sandi lama salah.")
		return
	}

	hash, err := auth.HashPassword(newPass)
	if err != nil {
		h.serverError(w, err)
		return
	}
	if err := h.Store.UpdateAdminPassword(username, hash); err != nil {
		h.serverError(w, err)
		return
	}

	h.renderAdmin(w, r, "settings", adminView{
		Title: "Pengaturan", User: username, Section: "settings", Settings: s,
		Success: "Kata sandi berhasil diubah. Silakan login kembali jika diperlukan.",
	})
}

func hasLetter(s string) bool {
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			return true
		}
	}
	return false
}

func hasDigit(s string) bool {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}

// ---------- Image Upload ----------

func (h *Handlers) ImageUpload(w http.ResponseWriter, r *http.Request) {
	const maxSize = 5 << 20 // 5 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxSize+512)

	if err := r.ParseMultipartForm(maxSize); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "File terlalu besar (maks. 5 MB)."})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "File tidak ditemukan dalam permintaan."})
		return
	}
	defer file.Close()

	sniff := make([]byte, 512)
	n, _ := file.Read(sniff)
	ct := http.DetectContentType(sniff[:n])
	if _, err := file.Seek(0, 0); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Gagal memproses file."})
		return
	}

	extFor := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/webp": ".webp",
	}
	ext, ok := extFor[ct]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Format tidak didukung. Gunakan JPG, PNG, GIF, atau WebP."})
		return
	}

	base := sanitizeFilename(strings.TrimSuffix(filepath.Base(header.Filename), filepath.Ext(header.Filename)))
	if base == "" {
		base = "product"
	}
	filename := fmt.Sprintf("%s-%d%s", base, time.Now().UnixMilli(), ext)

	dest := filepath.Join(h.WebDir, "static", "images", filename)
	out, err := os.Create(dest)
	if err != nil {
		log.Printf("image upload: create %q: %v", dest, err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Gagal menyimpan file di server."})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		log.Printf("image upload: write %q: %v", dest, err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Gagal menyimpan file di server."})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"filename": filename})
}

func sanitizeFilename(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('-')
		}
	}
	return b.String()
}

// ---------- form helpers ----------

func pathID(r *http.Request) int64 {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id
}

func formInt(r *http.Request, key string) int {
	n, _ := strconv.Atoi(r.FormValue(key))
	return n
}

func formBool(r *http.Request, key string) bool {
	v := r.FormValue(key)
	return v == "on" || v == "1" || v == "true"
}

func parseIDList(raw string) []int64 {
	parts := strings.Split(raw, ",")
	out := make([]int64, 0, len(parts))
	for _, p := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		out = append(out, id)
	}
	return out
}
