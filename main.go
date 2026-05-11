// Package main launches the packing-list HTTP server.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/bejl/packing-list/internal/auth"
	"github.com/bejl/packing-list/internal/catalog"
	"github.com/bejl/packing-list/internal/config"
	pdb "github.com/bejl/packing-list/internal/db"
	"github.com/bejl/packing-list/internal/trash"
	"github.com/bejl/packing-list/internal/trips"
	"github.com/bejl/packing-list/internal/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		logger.Error("config", "err", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		logger.Error("mkdir data", "err", err)
		os.Exit(1)
	}
	secret, err := auth.LoadOrCreateSecret(cfg.DataDir, cfg.SessionSecret)
	if err != nil {
		logger.Error("session secret", "err", err)
		os.Exit(1)
	}

	db, err := pdb.Open(filepath.Join(cfg.DataDir, "data.db"))
	if err != nil {
		logger.Error("db open", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	users := auth.NewUsers(db)

	// Bootstrap: ensure BOOTSTRAP_EMAIL exists if provided.
	if cfg.BootstrapEmail != "" {
		if _, _, err := users.FindOrCreate(context.Background(), cfg.BootstrapEmail); err != nil {
			logger.Warn("bootstrap user", "err", err)
		}
	}

	now := func() time.Time { return time.Now() }
	mailer := auth.NewMailer(cfg, logger)

	server := &web.Server{
		Cfg:       cfg,
		Logger:    logger,
		Users:     users,
		Sessions:  auth.NewSessions(db, secret, now),
		Magic:     auth.NewMagic(db, mailer, cfg.BaseURL, now),
		RateLimit: auth.NewRateLimiter(10, time.Hour, now),
		Items:     catalog.NewItems(db),
		Bundles:   catalog.NewBundles(db),
		Trips:     trips.NewTrips(db),
		Sources:   trips.NewSources(db),
		Pack:      trips.NewPack(db),
		Renderer2: trips.NewRenderer(db),
		IsDev:     !cfg.SMTP.Configured(),
		Now:       now,
	}
	renderer, err := web.NewRenderer()
	if err != nil {
		logger.Error("renderer", "err", err)
		os.Exit(1)
	}
	server.Renderer = renderer
	server.Trash = trash.NewView(db, server.Items, server.Bundles, server.Trips)

	hsvr := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Periodic cleanup: expired magic tokens + sessions every hour.
	go func() {
		t := time.NewTicker(time.Hour)
		defer t.Stop()
		for range t.C {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			_ = server.Magic.PurgeExpired(ctx)
			_ = server.Sessions.PurgeExpired(ctx)
			cancel()
		}
	}()

	logger.Info("listening", "addr", hsvr.Addr, "baseURL", cfg.BaseURL)
	go func() {
		if err := hsvr.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	logger.Info("shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = hsvr.Shutdown(ctx)
	_ = strings.TrimSpace // keep import alive if unused after refactors
}
