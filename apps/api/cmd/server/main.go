package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cachij/write-me/apps/api/internal/modules/application_specs"
	"github.com/cachij/write-me/apps/api/internal/modules/assets"
	"github.com/cachij/write-me/apps/api/internal/modules/identity"
	"github.com/cachij/write-me/apps/api/internal/modules/writing_sessions"
	"github.com/cachij/write-me/apps/api/internal/platform/bridge"
	"github.com/cachij/write-me/apps/api/internal/platform/config"
	"github.com/cachij/write-me/apps/api/internal/platform/db"
	httpserver "github.com/cachij/write-me/apps/api/internal/platform/http"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	store, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer store.Close()

	if err := store.Migrate(ctx); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	identityService := identity.NewService(store)
	if err := identityService.BootstrapAdmin(ctx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Fatalf("failed to bootstrap admin user: %v", err)
	}

	assetService := assets.NewService(store)
	specService := application_specs.NewService(store)
	bridgeClient := bridge.NewClient(cfg)
	bridgeProvider := bridge.NewAssistantProvider(bridgeClient)
	sessionService := writing_sessions.NewService(store, bridgeProvider)

	handler := httpserver.NewRouter(cfg, identityService, assetService, specService, sessionService, bridgeProvider)
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("write-me api listening on :%d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
