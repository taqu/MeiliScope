package main

import (
	"embed"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"

	appcore "github.com/taqu/meiliscope/internal/app"
	"github.com/taqu/meiliscope/internal/config"
	"github.com/taqu/meiliscope/internal/platform"
	"github.com/taqu/meiliscope/internal/services"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	cfgManager, err := config.NewManager()
	if err != nil {
		slog.Error("failed to initialise config manager", "err", err)
		os.Exit(1)
	}

	logDir := filepath.Join(cfgManager.Dir(), "logs")
	logCloser, _ := appcore.SetupLogging(logDir, slog.LevelInfo)
	if logCloser != nil {
		defer logCloser.Close()
	}

	if err := cfgManager.Load(); err != nil {
		slog.Error("failed to load config", "err", err)
		// Continue with defaults; non-fatal.
	}

	cfg := cfgManager.Get()

	creds, credErr := platform.NewCredentialStore()
	if credErr != nil {
		slog.Warn("credential store degraded; session-only mode active", "err", credErr)
	}

	appSvc := services.NewAppService()
	connSvc := services.NewConnectionService(cfgManager, creds)
	idxSvc := services.NewIndexService(connSvc)
	docSvc := services.NewDocumentService(connSvc)
	querySvc := services.NewQueryService(connSvc, cfgManager)

	app := application.New(application.Options{
		Name:        "MeiliScope",
		Description: "Desktop browser for Meilisearch",
		Services: []application.Service{
			application.NewService(appSvc),
			application.NewService(connSvc),
			application.NewService(idxSvc),
			application.NewService(docSvc),
			application.NewService(querySvc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "MeiliScope",
		Width:            cfg.WindowWidth,
		Height:           cfg.WindowHeight,
		BackgroundColour: application.NewRGB(17, 24, 39),
		URL:              "/",
		MinWidth:         800,
		MinHeight:        600,
	})

	slog.Info("starting MeiliScope", "version", services.AppVersion())

	if err := app.Run(); err != nil {
		slog.Error("application exited with error", "err", err)
		os.Exit(1)
	}

	if err := cfgManager.Save(); err != nil {
		slog.Error("failed to save config on exit", "err", err)
	}
}
