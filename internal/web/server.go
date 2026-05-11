package web

import (
	"embed"
	"log/slog"
	"net/http"
	"time"

	"github.com/bejl/packing-list/internal/auth"
	"github.com/bejl/packing-list/internal/catalog"
	"github.com/bejl/packing-list/internal/config"
	"github.com/bejl/packing-list/internal/trash"
	"github.com/bejl/packing-list/internal/trips"
)

// Server bundles all dependencies for HTTP handlers.
type Server struct {
	Cfg       config.Config
	Logger    *slog.Logger
	Renderer  *Renderer
	Users     *auth.Users
	Sessions  *auth.Sessions
	Magic     *auth.Magic
	RateLimit *auth.RateLimiter
	Items     *catalog.Items
	Bundles   *catalog.Bundles
	Trips     *trips.Trips
	Sources   *trips.Sources
	Pack      *trips.Pack
	Renderer2 *trips.Renderer
	Trash     *trash.View
	IsDev     bool
	Now       func() time.Time
}

// Handler builds the http.Handler with routes + middleware stack.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Serve static files from embedded FS
	mux.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.FS(staticFS))))

	mux.HandleFunc("GET /login", s.getLogin)
	mux.HandleFunc("POST /login", s.postLogin)
	mux.HandleFunc("GET /auth/verify", s.getVerify)
	mux.HandleFunc("POST /logout", s.postLogout)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200); w.Write([]byte("ok")) })

	authed := http.NewServeMux()
	authed.HandleFunc("GET /{$}", s.getTripsIndex)

	authed.HandleFunc("GET /trips/new", s.getTripNew)
	authed.HandleFunc("POST /trips", s.postTripCreate)
	authed.HandleFunc("GET /trips/{id}", s.getTripDetail)
	authed.HandleFunc("PATCH /trips/{id}", s.patchTrip)
	authed.HandleFunc("DELETE /trips/{id}", s.deleteTrip)
	authed.HandleFunc("POST /trips/{id}/bundles", s.attachBundle)
	authed.HandleFunc("DELETE /trips/{id}/bundles/{bid}", s.detachBundle)
	authed.HandleFunc("POST /trips/{id}/extras", s.addExtra)
	authed.HandleFunc("PATCH /trips/{id}/items/{iid}", s.overrideItem)
	authed.HandleFunc("POST /trips/{id}/pack/{iid}", s.togglePack)
	authed.HandleFunc("POST /trips/{id}/members", s.inviteMember)
	authed.HandleFunc("DELETE /trips/{id}/members/{uid}", s.removeMember)

	authed.HandleFunc("GET /items", s.listItems)
	authed.HandleFunc("POST /items", s.createItem)
	authed.HandleFunc("PATCH /items/{id}", s.updateItem)
	authed.HandleFunc("DELETE /items/{id}", s.deleteItem)

	authed.HandleFunc("GET /bundles", s.listBundles)
	authed.HandleFunc("POST /bundles", s.createBundle)
	authed.HandleFunc("GET /bundles/{id}", s.editBundle)
	authed.HandleFunc("PATCH /bundles/{id}", s.updateBundle)
	authed.HandleFunc("DELETE /bundles/{id}", s.deleteBundle)
	authed.HandleFunc("POST /bundles/{id}/items", s.bundleAddItem)
	authed.HandleFunc("DELETE /bundles/{id}/items/{iid}", s.bundleRemoveItem)
	authed.HandleFunc("POST /bundles/{id}/children", s.bundleAddChild)
	authed.HandleFunc("DELETE /bundles/{id}/children/{cid}", s.bundleRemoveChild)

	authed.HandleFunc("GET /trash", s.getTrash)
	authed.HandleFunc("POST /trash/{kind}/{id}/restore", s.restore)
	authed.HandleFunc("DELETE /trash/{kind}/{id}", s.purge)

	authed.HandleFunc("GET /export", s.export)
	authed.HandleFunc("POST /import", s.importJSON)

	mux.Handle("/", auth.RequireUser(s.Sessions)(authed))

	return s.accessLog(auth.CSRF(mux))
}

