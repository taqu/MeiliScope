package app

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// SetupLogging initialises structured logging to both stderr and a per-session
// log file in the application data directory. Returns the log file for closing
// on shutdown; a nil closer is safe to call.
func SetupLogging(logDir string, level slog.Level) (io.Closer, error) {
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return nil, fmt.Errorf("app: create log dir: %w", err)
	}

	ts := time.Now().Format("2006-01-02T15-04-05")
	logPath := filepath.Join(logDir, fmt.Sprintf("meiliscope-%s.log", ts))
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		// Fall back to stderr only; do not abort startup.
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
		return nil, nil //nolint:nilerr
	}

	w := io.MultiWriter(os.Stderr, f)
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
	slog.Info("log file opened", "path", logPath)

	return f, nil
}
