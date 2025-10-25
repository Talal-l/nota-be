package service

import (
	"context"
	"fmt"
	"io"
	"nota/service/db"
	"nota/types"

	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/uptrace/bun"

	"github.com/anotik/anocore/pkg/logger"
)

// @title           Nota API
// @version         1.0
// @description     API documentation for the Nota backend.
// @BasePath        /nota/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name x-api-key
// @description Custom API Key header

// @securityDefinitions.apikey ClientIdAuth
// @in header
// @name x-client-id
// @description Custom Client ID header

// @security ApiKeyAuth
// @security ClientIdAuth

// Handles staring and shutting down the server
// this is used to help us start consistent instance of the server when testing. We also get the change to do cleanup logic when we receive a shutdown signal
func Run(
	ctx context.Context,
	cfg types.AppConfig,
	args []string,
	stdin io.Reader,
	stdout, stderr io.Writer,
) error {

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	log, err := logger.NewContextLogger(ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to create context logger: %v", err))
	}

	log.LineInfo = false
	log.Info("config", "config", cfg)

	db, err := db.NewDB(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create database connection: %v", err))
	}

	srv := NewServer(
		log,
		cfg,
		db,
	)

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(cfg.Host, cfg.Port),
		Handler: srv,
	}

	var shutdownErr error
	var shutdownReason error

	errChan := make(chan error, 1)
	go func() {
		log.Info("Starting Service", "port", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
			cancel()
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			shutdownReason = ctx.Err()
		case err := <-errChan:
			shutdownReason = err
		}
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		log.Info("Shutting down Service", "port", cfg.Port, "reason", shutdownReason)
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error("Failed to shutdown Service", "port", cfg.Port, "error", err)
			shutdownErr = err
		}
	}()
	wg.Wait()
	if shutdownErr != nil {
		return shutdownErr
	}
	return nil
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Help us easily create the full server setup with all middleware and makes our dep more explicit
func NewServer(logger *logger.ContextLogger, cnf types.AppConfig, db *bun.DB) http.Handler {

	mux := http.NewServeMux()
	addRoutes(mux, cnf, db)

	if cnf.Mode == "dev" {

		return withCORS(mux)
	}
	return mux
}

// waitForReady calls the specified endpoint until it gets a 200
// response or until the context is cancelled or the timeout is
// reached.
func WaitForReady(
	ctx context.Context,
	log *logger.ContextLogger,
	timeout time.Duration,
	endpoint string,
) error {
	client := http.Client{}
	startTime := time.Now()
	for {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			endpoint,
			nil,
		)
		if err != nil {
			log.Error("Error creating request", "error", err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Error("Error making request", "error", err)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			log.Info("Endpoint is ready", "endpoint", endpoint)
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(startTime) >= timeout {
				return fmt.Errorf("timeout reached while waiting for endpoint")
			}
			// wait a little while between checks
			time.Sleep(250 * time.Millisecond)
		}
	}
}
