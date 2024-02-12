package app

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
	"github.com/apoldev/go-http/pkg/logger"
)

type App struct {
	srv    *http.Server
	logger logger.Logger
}

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

func New() (*App, error) {
	addr := env.LookupEnvStringDefault("ADDR", DefaultAddr)
	maxConnections := env.LookupEnvIntDefault("SERVER_MAX_CONNECTIONS", DefaultMaxConnections)
	maxUrlsCount := env.LookupEnvIntDefault("CRAWLER_MAX_URLS", DefaultMaxUrlsCount)
	maxWorkersCount := env.LookupEnvIntDefault("CRAWLER_MAX_WORKERS", DefaultMaxWorkers)
	crawlerRequestTimeoutMs := env.LookupEnvIntDefault("CRAWLER_REQUEST_TIMEOUT_MS", DefaultCrawlerRequestTimeoutMs)

	limiter := limiter.NewAtomLimiter(maxConnections)

	// todo add proxy to client Transport
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
	srv := &http.Server{
		Addr:        addr,
		Handler:     mux,
		IdleTimeout: DefaultServerIdleTimeout,
		ReadTimeout: DefaultServerReadWriteTimeout,
	}

	return &App{
		logger: log.New(os.Stdout, "[main] ", log.LstdFlags),
		srv:    srv,
	}, nil
}

func (a *App) Run() error {
	go func() {
		err := a.srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Printf("Server failed: %v", err)
		}
	}()

	a.logger.Printf("Server started on %s", a.srv.Addr)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done

	a.logger.Printf("Server stopping")

	ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	if err := a.srv.Shutdown(ctx); err != nil {
		return err
	}
	a.logger.Printf("Server stopped")
	return nil
}
