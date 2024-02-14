package crawler

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/apoldev/go-http/pkg/logger"
)

type Service struct {
	workerCount    int
	requestTimeout time.Duration
	httpClient     *http.Client
	logger         logger.Logger
}

func New(workerCount, crawlerRequestTimeoutMs int, httpClient *http.Client, logger logger.Logger) *Service {
	return &Service{
		workerCount:    workerCount,
		httpClient:     httpClient,
		logger:         logger,
		requestTimeout: time.Millisecond * time.Duration(crawlerRequestTimeoutMs),
	}
}

type resultCrawl struct {
	Data []byte
	URL  string
	Err  error
}

// Crawl is a method for crawling multiple URLs.
func (c *Service) Crawl(ctx context.Context, urls []string) (map[string][]byte, error) {
	ch := make(chan string, len(urls))
	resultCh := make(chan resultCrawl)
	var wg sync.WaitGroup
	var cancel context.CancelFunc

	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	wg.Add(c.workerCount)
	for i := 0; i < c.workerCount; i++ {
		go c.worker(ctx, ch, resultCh, &wg)
	}

	for i := range urls {
		ch <- urls[i]
	}
	close(ch)

	go func() {
		wg.Wait()
		c.logger.Print("all workers are finished")
		close(resultCh)
	}()

	results := make(map[string][]byte)
	for res := range resultCh {
		if res.Err != nil {
			c.logger.Printf("got error at %s. Error: %v", res.URL, res.Err)
			return nil, res.Err
		}
		c.logger.Printf("got data from %s. Content-Length: %d", res.URL, len(res.Data))
		results[res.URL] = res.Data
	}

	return results, nil
}

func (c *Service) worker(ctx context.Context, ch <-chan string, resultCh chan<- resultCrawl, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		select {
		case <-ctx.Done():
			return
		case url, ok := <-ch:
			if !ok {
				return
			}
			data, err := c.httpRequest(ctx, url)
			if err != nil {
				resultCh <- resultCrawl{
					Err: err,
					URL: url,
				}
				return
			}
			resultCh <- resultCrawl{
				Data: data,
				URL:  url,
			}
		}
	}
}

func (c *Service) httpRequest(ctx context.Context, link string) ([]byte, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, err
}
