package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	httpresp "github.com/apoldev/go-http/internal/app/lib/http-resp"
	"github.com/apoldev/go-http/pkg/logger"
)

//go:generate go run github.com/vektra/mockery/v2@v2.40.1 --name Service
type Service interface {
	Crawl(ctx context.Context, urls []string) (map[string][]byte, error)
}

// HTTPHandler is a handler for http request
type HTTPHandler struct {
	crawlService Service
	maxUrls      int
	logger       logger.Logger
}

func NewHTTPHandler(crawlService Service, maxUrls int, logger logger.Logger) *HTTPHandler {
	return &HTTPHandler{
		crawlService: crawlService,
		maxUrls:      maxUrls,
		logger:       logger,
	}
}

type CrawlRequest []string
type CrawlResponse map[string]string

// Crawl is a handler for http request that helps crawl multiple URLs.
func (h *HTTPHandler) Crawl(w http.ResponseWriter, r *http.Request) {
	var err error
	ctx := r.Context()

	if r.Method != http.MethodPost {
		httpresp.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// json
	var urls CrawlRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&urls)
	if err != nil {
		httpresp.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// validate count of urls
	if len(urls) > h.maxUrls {
		httpresp.Error(w, fmt.Sprintf("Too many urls. Max is %d", h.maxUrls), http.StatusBadRequest)
		return
	}

	// call crawl()
	data, err := h.crawlService.Crawl(ctx, urls)
	if errors.Is(err, context.Canceled) {
		httpresp.Error(w, fmt.Sprintf("request canceled: %s", err), http.StatusInternalServerError)
		return
	}
	if err != nil {
		httpresp.Error(w, fmt.Sprintf("Internal Server Error: %s", err), http.StatusInternalServerError)
		return
	}

	resp := make(CrawlResponse)
	for k, v := range data {
		resp[k] = string(v)
	}
	httpresp.WriteJSON(w, resp, http.StatusOK)
}
