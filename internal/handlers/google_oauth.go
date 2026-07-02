package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	googleAuthURL     = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL    = "https://oauth2.googleapis.com/token"
	googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
)

// GoogleLoginStart redirects the user to Google's OAuth consent screen.
func (h *Handlers) GoogleLoginStart(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		h.renderAdmin(w, r, "login", adminView{
			Title: "Masuk Admin",
			Error: "Google login belum dikonfigurasi (GOOGLE_CLIENT_ID tidak diset).",
		})
		return
	}

	b := make([]byte, 16)
	rand.Read(b)
	state := hex.EncodeToString(b)

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		MaxAge:   300,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", googleRedirectURI(r))
	params.Set("response_type", "code")
	params.Set("scope", "openid email profile")
	params.Set("state", state)
	params.Set("access_type", "online")
	params.Set("prompt", "select_account")

	http.Redirect(w, r, googleAuthURL+"?"+params.Encode(), http.StatusTemporaryRedirect)
}

// GoogleCallback handles the OAuth2 callback from Google.
func (h *Handlers) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	allowedEmail := os.Getenv("GOOGLE_ALLOWED_EMAIL")

	// Validate CSRF state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.FormValue("state") {
		h.renderAdmin(w, r, "login", adminView{Title: "Masuk Admin", Error: "Permintaan tidak valid, coba lagi."})
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: "", MaxAge: -1, Path: "/"})

	if errParam := r.FormValue("error"); errParam != "" {
		h.renderAdmin(w, r, "login", adminView{Title: "Masuk Admin", Error: "Login Google dibatalkan."})
		return
	}

	code := r.FormValue("code")
	token, err := exchangeGoogleCode(clientID, clientSecret, code, googleRedirectURI(r))
	if err != nil {
		h.renderAdmin(w, r, "login", adminView{Title: "Masuk Admin", Error: "Gagal verifikasi token Google."})
		return
	}

	userInfo, err := getGoogleUserInfo(token.AccessToken)
	if err != nil || userInfo.Email == "" {
		h.renderAdmin(w, r, "login", adminView{Title: "Masuk Admin", Error: "Gagal mengambil info akun Google."})
		return
	}

	if allowedEmail != "" && !strings.EqualFold(userInfo.Email, allowedEmail) {
		h.renderAdmin(w, r, "login", adminView{
			Title: "Masuk Admin",
			Error: fmt.Sprintf("Akun %s tidak diizinkan mengakses admin.", userInfo.Email),
		})
		return
	}

	h.Sessions.Create(w, userInfo.Email)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func googleRedirectURI(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/admin/auth/google/callback", scheme, r.Host)
}

type googleTokenResp struct {
	AccessToken string `json:"access_token"`
}

type googleUserInfoResp struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func exchangeGoogleCode(clientID, clientSecret, code, redirectURI string) (*googleTokenResp, error) {
	params := url.Values{}
	params.Set("code", code)
	params.Set("client_id", clientID)
	params.Set("client_secret", clientSecret)
	params.Set("redirect_uri", redirectURI)
	params.Set("grant_type", "authorization_code")

	resp, err := http.PostForm(googleTokenURL, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var t googleTokenResp
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}
	if t.AccessToken == "" {
		return nil, fmt.Errorf("no access_token in response")
	}
	return &t, nil
}

func getGoogleUserInfo(accessToken string) (*googleUserInfoResp, error) {
	req, err := http.NewRequest("GET", googleUserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info googleUserInfoResp
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	return &info, nil
}
