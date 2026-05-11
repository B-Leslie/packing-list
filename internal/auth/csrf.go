package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
)

const (
	CSRFCookieName = "csrf"
	CSRFHeaderName = "X-CSRF-Token"
	CSRFFormField  = "csrf_token"
)

var safeMethods = map[string]bool{
	"GET": true, "HEAD": true, "OPTIONS": true,
}

// CSRF issues a CSRF cookie on safe methods and validates it on unsafe ones.
// The token is doubled in cookie + header/form; we compare with constant time.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(CSRFCookieName)
		if err != nil || c.Value == "" {
			var raw [24]byte
			rand.Read(raw[:])
			tok := base64.RawURLEncoding.EncodeToString(raw[:])
			http.SetCookie(w, &http.Cookie{
				Name:     CSRFCookieName,
				Value:    tok,
				Path:     "/",
				HttpOnly: false, // readable so HTMX can echo into header
				SameSite: http.SameSiteLaxMode,
			})
			c = &http.Cookie{Name: CSRFCookieName, Value: tok}
		}
		if safeMethods[r.Method] {
			next.ServeHTTP(w, r)
			return
		}
		got := r.Header.Get(CSRFHeaderName)
		if got == "" {
			// Form fallback. Parse body once; downstream code re-parses freely.
			if err := r.ParseForm(); err == nil {
				got = r.FormValue(CSRFFormField)
			}
		}
		if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(c.Value)) != 1 {
			http.Error(w, "csrf failure", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
