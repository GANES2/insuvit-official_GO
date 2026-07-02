package db

import (
	"database/sql"

	"insuvit/internal/models"
)

// Store wraps the database connection with typed query helpers.
type Store struct {
	DB *sql.DB
}

// NewStore returns a Store over the given connection.
func NewStore(conn *sql.DB) *Store { return &Store{DB: conn} }

// ---------- Settings ----------

func (s *Store) AllSettings() (models.Settings, error) {
	rows, err := s.DB.Query("SELECT key, value FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := models.Settings{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

func (s *Store) UpdateSettings(values map[string]string) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	for k, v := range values {
		if _, err := tx.Exec(
			`INSERT INTO settings (key, value) VALUES (?, ?)
			 ON CONFLICT(key) DO UPDATE SET value = excluded.value`, k, v); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// ---------- Products ----------

func (s *Store) ListProducts(activeOnly bool) ([]models.Product, error) {
	q := `SELECT id, name, description, meta_left, meta_right, image, shopee_url, badge, ribbon, sort_order, is_active FROM products`
	if activeOnly {
		q += " WHERE is_active = 1"
	}
	q += " ORDER BY sort_order, id"
	rows, err := s.DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.MetaLeft, &p.MetaRight, &p.Image, &p.ShopeeURL, &p.Badge, &p.Ribbon, &p.SortOrder, &p.IsActive); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GetProduct(id int64) (models.Product, error) {
	var p models.Product
	err := s.DB.QueryRow(
		`SELECT id, name, description, meta_left, meta_right, image, shopee_url, badge, ribbon, sort_order, is_active FROM products WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.MetaLeft, &p.MetaRight, &p.Image, &p.ShopeeURL, &p.Badge, &p.Ribbon, &p.SortOrder, &p.IsActive)
	return p, err
}

func (s *Store) CreateProduct(p models.Product) (int64, error) {
	res, err := s.DB.Exec(
		`INSERT INTO products (name, description, meta_left, meta_right, image, shopee_url, badge, ribbon, sort_order, is_active)
		 VALUES (?,?,?,?,?,?,?,?,?,?)`,
		p.Name, p.Description, p.MetaLeft, p.MetaRight, p.Image, p.ShopeeURL, p.Badge, p.Ribbon, p.SortOrder, p.IsActive)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpdateProduct(p models.Product) error {
	_, err := s.DB.Exec(
		`UPDATE products SET name=?, description=?, meta_left=?, meta_right=?, image=?, shopee_url=?, badge=?, ribbon=?, sort_order=?, is_active=? WHERE id=?`,
		p.Name, p.Description, p.MetaLeft, p.MetaRight, p.Image, p.ShopeeURL, p.Badge, p.Ribbon, p.SortOrder, p.IsActive, p.ID)
	return err
}

func (s *Store) DeleteProduct(id int64) error {
	_, err := s.DB.Exec("DELETE FROM products WHERE id = ?", id)
	return err
}

func (s *Store) ReorderProducts(ids []int64) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	for i, id := range ids {
		if _, err := tx.Exec("UPDATE products SET sort_order = ? WHERE id = ?", i*10, id); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// ---------- Testimonials ----------

func (s *Store) ListTestimonials(publishedOnly bool) ([]models.Testimonial, error) {
	q := `SELECT id, quote, author, role, sort_order, is_published FROM testimonials`
	if publishedOnly {
		q += " WHERE is_published = 1"
	}
	q += " ORDER BY sort_order, id"
	rows, err := s.DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Testimonial
	for rows.Next() {
		var t models.Testimonial
		if err := rows.Scan(&t.ID, &t.Quote, &t.Author, &t.Role, &t.SortOrder, &t.IsPublished); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) GetTestimonial(id int64) (models.Testimonial, error) {
	var t models.Testimonial
	err := s.DB.QueryRow(
		`SELECT id, quote, author, role, sort_order, is_published FROM testimonials WHERE id = ?`, id,
	).Scan(&t.ID, &t.Quote, &t.Author, &t.Role, &t.SortOrder, &t.IsPublished)
	return t, err
}

func (s *Store) CreateTestimonial(t models.Testimonial) (int64, error) {
	res, err := s.DB.Exec(
		`INSERT INTO testimonials (quote, author, role, sort_order, is_published) VALUES (?,?,?,?,?)`,
		t.Quote, t.Author, t.Role, t.SortOrder, t.IsPublished)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpdateTestimonial(t models.Testimonial) error {
	_, err := s.DB.Exec(
		`UPDATE testimonials SET quote=?, author=?, role=?, sort_order=?, is_published=? WHERE id=?`,
		t.Quote, t.Author, t.Role, t.SortOrder, t.IsPublished, t.ID)
	return err
}

func (s *Store) DeleteTestimonial(id int64) error {
	_, err := s.DB.Exec("DELETE FROM testimonials WHERE id = ?", id)
	return err
}

func (s *Store) ReorderTestimonials(ids []int64) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	for i, id := range ids {
		if _, err := tx.Exec("UPDATE testimonials SET sort_order = ? WHERE id = ?", i*10, id); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// ---------- FAQs ----------

func (s *Store) ListFAQs(activeOnly bool) ([]models.FAQ, error) {
	q := `SELECT id, question, answer, sort_order, is_active FROM faqs`
	if activeOnly {
		q += " WHERE is_active = 1"
	}
	q += " ORDER BY sort_order, id"
	rows, err := s.DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.FAQ
	for rows.Next() {
		var f models.FAQ
		if err := rows.Scan(&f.ID, &f.Question, &f.Answer, &f.SortOrder, &f.IsActive); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (s *Store) GetFAQ(id int64) (models.FAQ, error) {
	var f models.FAQ
	err := s.DB.QueryRow(
		`SELECT id, question, answer, sort_order, is_active FROM faqs WHERE id = ?`, id,
	).Scan(&f.ID, &f.Question, &f.Answer, &f.SortOrder, &f.IsActive)
	return f, err
}

func (s *Store) CreateFAQ(f models.FAQ) (int64, error) {
	res, err := s.DB.Exec(
		`INSERT INTO faqs (question, answer, sort_order, is_active) VALUES (?,?,?,?)`,
		f.Question, f.Answer, f.SortOrder, f.IsActive)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpdateFAQ(f models.FAQ) error {
	_, err := s.DB.Exec(
		`UPDATE faqs SET question=?, answer=?, sort_order=?, is_active=? WHERE id=?`,
		f.Question, f.Answer, f.SortOrder, f.IsActive, f.ID)
	return err
}

func (s *Store) DeleteFAQ(id int64) error {
	_, err := s.DB.Exec("DELETE FROM faqs WHERE id = ?", id)
	return err
}

func (s *Store) ReorderFAQs(ids []int64) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	for i, id := range ids {
		if _, err := tx.Exec("UPDATE faqs SET sort_order = ? WHERE id = ?", i*10, id); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// ---------- Admin ----------

func (s *Store) GetAdminByUsername(username string) (models.AdminUser, error) {
	var u models.AdminUser
	err := s.DB.QueryRow(
		`SELECT id, username, password_hash FROM admin_users WHERE username = ?`, username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash)
	return u, err
}

func (s *Store) UpdateAdminPassword(username, newHash string) error {
	_, err := s.DB.Exec(`UPDATE admin_users SET password_hash = ? WHERE username = ?`, newHash, username)
	return err
}

// ---------- Analytics ----------

func (s *Store) RecordPageView(path, userAgent, ip string) error {
	_, err := s.DB.Exec(
		`INSERT INTO page_views (path, user_agent, ip_address) VALUES (?, ?, ?)`,
		path, userAgent, ip,
	)
	return err
}

func (s *Store) RecordClickEvent(target, url, userAgent, ip string) error {
	_, err := s.DB.Exec(
		`INSERT INTO click_events (target, url, user_agent, ip_address) VALUES (?, ?, ?, ?)`,
		target, url, userAgent, ip,
	)
	return err
}

func (s *Store) GetAnalyticsStats() (map[string]int, error) {
	stats := map[string]int{
		"page_views":    0,
		"shopee_clicks": 0,
		"wa_clicks":     0,
	}

	var count int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM page_views WHERE path NOT LIKE '/admin%' AND path NOT LIKE '/static%'`).Scan(&count)
	if err != nil {
		return stats, err
	}
	stats["page_views"] = count

	err = s.DB.QueryRow(`SELECT COUNT(*) FROM click_events WHERE target = 'shopee'`).Scan(&count)
	if err != nil {
		return stats, err
	}
	stats["shopee_clicks"] = count

	err = s.DB.QueryRow(`SELECT COUNT(*) FROM click_events WHERE target = 'whatsapp'`).Scan(&count)
	if err != nil {
		return stats, err
	}
	stats["wa_clicks"] = count

	return stats, nil
}
