package web

import (
	"errors"
	"net/http"
	"strings"

	"github.com/bejl/packing-list/internal/auth"
)

func (s *Server) getLogin(w http.ResponseWriter, r *http.Request) {
	s.Renderer.Render(w, "login", map[string]any{"Title": "Sign in"})
}

func (s *Server) postLogin(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" {
		s.Renderer.Render(w, "login", map[string]any{"Title": "Sign in", "Error": "Email is required"})
		return
	}
	keyEmail := strings.ToLower(email)
	keyIP := r.RemoteAddr
	if !s.RateLimit.Allow(keyEmail) || !s.RateLimit.Allow(keyIP) {
		s.Renderer.Render(w, "login", map[string]any{"Title": "Sign in", "Error": "Too many attempts. Try again later."})
		return
	}
	// Invite-only: never auto-create on /login. Unknown emails still get
	// the "check your inbox" page so the response shape doesn't leak
	// which addresses are allowed. No magic link is actually issued.
	_, found, err := s.Users.Find(r.Context(), email)
	if err != nil {
		s.Logger.Error("find user", "err", err)
		s.Renderer.Render(w, "login", map[string]any{"Title": "Sign in", "Error": "Could not start sign-in"})
		return
	}
	if found {
		if err := s.Magic.Issue(r.Context(), email); err != nil {
			s.Logger.Error("issue magic link", "err", err)
			s.Renderer.Render(w, "login", map[string]any{"Title": "Sign in", "Error": "Could not send sign-in email"})
			return
		}
	} else {
		s.Logger.Info("login attempt for unknown email", "email", email)
	}
	data := map[string]any{"Title": "Check inbox", "Email": email}
	if s.IsDev && found {
		data["DevLink"] = "(see server log)"
	}
	s.Renderer.Render(w, "login_sent", data)
}

func (s *Server) getVerify(w http.ResponseWriter, r *http.Request) {
	tok := r.URL.Query().Get("t")
	email, err := s.Magic.Consume(r.Context(), tok)
	if err != nil {
		s.Renderer.Render(w, "login", map[string]any{"Title": "Sign in", "Error": "Link is invalid or expired"})
		return
	}
	// Defense-in-depth: even if a token was minted for an unknown email
	// (shouldn't happen -- postLogin gates on Find), require the user
	// to exist at consumption time.
	uid, found, err := s.Users.Find(r.Context(), email)
	if err != nil || !found {
		s.Logger.Error("find on verify", "err", err, "found", found)
		s.Renderer.Render(w, "login", map[string]any{"Title": "Sign in", "Error": "Could not finish sign-in"})
		return
	}
	cookieVal, err := s.Sessions.Issue(r.Context(), uid)
	if err != nil {
		s.Logger.Error("issue session", "err", err)
		s.Renderer.Render(w, "login", map[string]any{"Title": "Sign in", "Error": "Could not start session"})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    cookieVal,
		Path:     "/",
		HttpOnly: true,
		Secure:   strings.HasPrefix(s.Cfg.BaseURL, "https://"),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(auth.SessionTTL.Seconds()),
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) postLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(auth.CookieName); err == nil {
		_ = s.Sessions.Revoke(r.Context(), c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: auth.CookieName, Value: "", MaxAge: -1, Path: "/"})
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// accessLog logs method + path + status + duration.
func (s *Server) accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &statusRW{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		s.Logger.Info("http", "method", r.Method, "path", r.URL.Path, "status", rw.status)
	})
}

type statusRW struct {
	http.ResponseWriter
	status int
}

func (s *statusRW) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

var _ = errors.New
