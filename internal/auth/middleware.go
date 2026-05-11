package auth

import (
	"context"
	"net/http"
)

const CookieName = "sid"

type ctxKey string

const userKey ctxKey = "uid"

func UserFrom(ctx context.Context) string {
	v, _ := ctx.Value(userKey).(string)
	return v
}

func WithUser(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, userKey, uid)
}

func RequireUser(s *Sessions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(CookieName)
			if err != nil {
				unauthorized(w, r)
				return
			}
			uid, err := s.Lookup(r.Context(), c.Value)
			if err != nil {
				// Clear bad cookie so the browser stops sending it.
				http.SetCookie(w, &http.Cookie{Name: CookieName, Value: "", MaxAge: -1, Path: "/"})
				unauthorized(w, r)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), uid)))
		})
	}
}

func unauthorized(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
