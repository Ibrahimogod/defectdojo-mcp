package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ibrahimogod/defectdojo-mcp/internal/config"
	"github.com/ibrahimogod/defectdojo-mcp/internal/dojoclient"
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
	client := dojoclient.New(cfg.DojoBaseURL, cfg.RequestTimeout)

	var fallbackAuth string
	if cfg.APIToken != "" {
		fallbackAuth = "Token " + cfg.APIToken
	}

	server := mcpserver.New(client, fallbackAuth, cfg.EnableDestructive)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.Transport == config.TransportStdio {
		if err := mcpserver.RunStdio(ctx, server); err != nil {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
		return
	}

	httpSrv := mcpserver.NewHTTPServer(cfg.ListenAddr, server, cfg.DojoBaseURL, cfg.RequestTimeout)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", httpSrv.Addr())
		errCh <- httpSrv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("shutting down")
		if err := httpSrv.Shutdown(context.Background()); err != nil {
			logger.Error("shutdown error", "error", err)
			os.Exit(1)
		}
	}
}
