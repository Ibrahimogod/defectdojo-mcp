package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ibrahimogod/defectdojo-mcp/internal/config"
	applog "github.com/ibrahimogod/defectdojo-mcp/internal/log"
	"github.com/ibrahimogod/defectdojo-mcp/internal/mcpserver"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}

	logger := applog.New(cfg.LogLevel)
	srv := mcpserver.New(cfg.ListenAddr, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", srv.Addr())
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("shutting down")
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Error("shutdown error", "error", err)
			os.Exit(1)
		}
	}
}
