package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/apoldev/go-http/internal/app/crawler"
	"github.com/apoldev/go-http/internal/app/handlers"
	"github.com/apoldev/go-http/internal/app/lib/env"
	"github.com/apoldev/go-http/internal/app/limiter"
	"github.com/apoldev/go-http/internal/app/middleware"
)

const (
	DefaultMaxConnections          = 1
	DefaultMaxUrlsCount            = 20
	DefaultMaxWorkers              = 4
	DefaultAddr                    = ":8080"
	DefaultCrawlerRequestTimeoutMs = 1000
	DefaultServerReadWriteTimeout  = time.Second * 10
	DefaultServerIdleTimeout       = time.Second * 60
	DefaultShutdownTimeout         = time.Second * 15
)

func main() {
	addr := env.LookupEnvStringDefault("ADDR", DefaultAddr)
	maxConnections := env.LookupEnvIntDefault("SERVER_MAX_CONNECTIONS", DefaultMaxConnections)
	maxUrlsCount := env.LookupEnvIntDefault("CRAWLER_MAX_URLS", DefaultMaxUrlsCount)
	maxWorkersCount := env.LookupEnvIntDefault("CRAWLER_MAX_WORKERS", DefaultMaxWorkers)
	crawlerRequestTimeoutMs := env.LookupEnvIntDefault("CRAWLER_REQUEST_TIMEOUT_MS", DefaultCrawlerRequestTimeoutMs)

	limiter := limiter.NewAtomLimiter(maxConnections)

	logger := log.New(os.Stdout, "[main] ", log.LstdFlags)

	httpClient := http.DefaultClient
	crawleService := crawler.New(
		maxWorkersCount,
		crawlerRequestTimeoutMs,
		httpClient,
		log.New(os.Stdout, "[crawler] ", log.LstdFlags),
	)
	httpHandler := handlers.NewHTTPHandler(
		crawleService,
		maxUrlsCount,
		log.New(os.Stdout, "[http] ", log.LstdFlags),
	)

	mux := http.NewServeMux()

	handler := middleware.LimitMiddleware(limiter, http.HandlerFunc(httpHandler.Crawl))
	mux.Handle("/", handler)

	server := http.Server{
		Addr:        addr,
		Handler:     mux,
		IdleTimeout: DefaultServerIdleTimeout,
		ReadTimeout: DefaultServerReadWriteTimeout,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Printf("Server failed: %v", err)
		}
	}()

	logger.Printf("Server started on %s", addr)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done

	logger.Printf("Server stopping")

	ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("failed to stop server: %v", err)
		return
	}

	logger.Printf("Server stopped")
}
