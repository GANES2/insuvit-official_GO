// Package auth provides password hashing and signed cookie sessions
// for the admin panel using only the Go standard library.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const sessionCookie = "insuvit_session"

// HashPassword returns a "saltHex:hashHex" string for storage.
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	sum := sha256.Sum256(append(salt, []byte(password)...))
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(sum[:]), nil
}

// CheckPassword reports whether password matches a stored "saltHex:hashHex".
func CheckPassword(password, stored string) bool {
	parts := strings.SplitN(stored, ":", 2)
	if len(parts) != 2 {
		return false
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}
	want, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	sum := sha256.Sum256(append(salt, []byte(password)...))
	return subtle.ConstantTimeCompare(sum[:], want) == 1
}

// SessionManager signs and verifies session cookies with an HMAC secret.
type SessionManager struct {
	secret []byte
	maxAge time.Duration
	secure bool
}

// NewSessionManager creates a manager. secret should be kept stable across
// restarts (set SESSION_SECRET) so existing logins remain valid.
func NewSessionManager(secret string) *SessionManager {
	return &SessionManager{
		secret: []byte(secret),
		maxAge: 8 * time.Hour,
		secure: false, // set true behind HTTPS
	}
}

func (s *SessionManager) sign(payload string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// Create issues a signed session cookie for the given username.
func (s *SessionManager) Create(w http.ResponseWriter, username string) {
	exp := time.Now().Add(s.maxAge).Unix()
	payload := fmt.Sprintf("%s|%d", base64.RawURLEncoding.EncodeToString([]byte(username)), exp)
	value := payload + "|" + s.sign(payload)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(s.maxAge.Seconds()),
	})
}

// Validate returns the logged-in username, or an error if the session is
// missing, tampered with, or expired.
func (s *SessionManager) Validate(r *http.Request) (string, error) {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return "", errors.New("no session")
	}
	parts := strings.Split(c.Value, "|")
	if len(parts) != 3 {
		return "", errors.New("malformed session")
	}
	payload := parts[0] + "|" + parts[1]
	if !hmac.Equal([]byte(s.sign(payload)), []byte(parts[2])) {
		return "", errors.New("bad signature")
	}
	exp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return "", errors.New("expired")
	}
	nameBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", errors.New("bad payload")
	}
	return string(nameBytes), nil
}

// Destroy clears the session cookie (logout).
func (s *SessionManager) Destroy(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// CSRFToken derives a CSRF token bound to the current session cookie.
// The token is unique per session and cannot be forged without the secret.
func (s *SessionManager) CSRFToken(r *http.Request) string {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		b := make([]byte, 16)
		rand.Read(b) //nolint:errcheck
		return hex.EncodeToString(b)
	}
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte("csrf:" + c.Value))
	return hex.EncodeToString(mac.Sum(nil))
}

// ValidateCSRF checks that the "csrf_token" form field matches the expected
// token for the current session. Must be called after ParseForm.
func (s *SessionManager) ValidateCSRF(r *http.Request) bool {
	formToken := r.FormValue("csrf_token")
	if formToken == "" {
		return false
	}
	expected := s.CSRFToken(r)
	return hmac.Equal([]byte(formToken), []byte(expected))
}

// ---------------------------------------------------------------------------
// Login rate limiter
// ---------------------------------------------------------------------------

const (
	maxLoginFailures = 5 // failures before blocking
	loginBlockFor    = 15 * time.Minute
	loginFailWindow  = 10 * time.Minute
)

type loginRecord struct {
	count     int
	lastFail  time.Time
	blockedAt time.Time
}

// LoginLimiter is a simple in-memory rate limiter for login attempts, keyed
// by IP address. It blocks an IP for loginBlockFor after maxLoginFailures
// consecutive failures within loginFailWindow.
type LoginLimiter struct {
	mu      sync.Mutex
	records map[string]*loginRecord
}

// NewLoginLimiter creates a LoginLimiter ready for use.
func NewLoginLimiter() *LoginLimiter {
	return &LoginLimiter{records: make(map[string]*loginRecord)}
}

// Allow returns true if the IP may attempt a login right now.
func (l *LoginLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	rec, ok := l.records[ip]
	if !ok {
		return true
	}
	if !rec.blockedAt.IsZero() {
		if time.Since(rec.blockedAt) < loginBlockFor {
			return false
		}
		delete(l.records, ip)
	}
	return true
}

// RecordFailure increments the failure count for ip and blocks it if the
// threshold is reached within the window.
func (l *LoginLimiter) RecordFailure(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	rec, ok := l.records[ip]
	if !ok {
		l.records[ip] = &loginRecord{count: 1, lastFail: time.Now()}
		return
	}
	if time.Since(rec.lastFail) > loginFailWindow {
		rec.count = 0
		rec.blockedAt = time.Time{}
	}
	rec.count++
	rec.lastFail = time.Now()
	if rec.count >= maxLoginFailures {
		rec.blockedAt = time.Now()
	}
}

// RecordSuccess clears the failure record for ip after a successful login.
func (l *LoginLimiter) RecordSuccess(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.records, ip)
}

// RemainingBlock returns how long ip is still blocked (zero if not blocked).
func (l *LoginLimiter) RemainingBlock(ip string) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	rec, ok := l.records[ip]
	if !ok || rec.blockedAt.IsZero() {
		return 0
	}
	remaining := loginBlockFor - time.Since(rec.blockedAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}
