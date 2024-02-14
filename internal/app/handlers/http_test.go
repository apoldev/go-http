package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apoldev/go-http/internal/app/handlers"
	"github.com/apoldev/go-http/internal/app/handlers/mocks"
	"github.com/stretchr/testify/require"
)

func TestCrawlHandler(t *testing.T) {
	logger := log.New(io.Discard, "", log.LstdFlags)

	cases := []struct {
		name            string
		body            []byte
		method          string
		needCallCrawler bool
		urls            []string
		expectedStatus  int
		results         map[string][]byte
		expectError     error
	}{
		{
			name:           "bad_method",
			method:         http.MethodGet,
			body:           []byte{},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:            "json_err",
			method:          http.MethodPost,
			body:            []byte{2, 32, 65},
			needCallCrawler: false,
			expectedStatus:  http.StatusBadRequest,
		},
		{
			name:            "valid request",
			method:          http.MethodPost,
			body:            []byte(`["https://google.com"]`),
			needCallCrawler: true,
			urls:            []string{"https://google.com"},
			expectedStatus:  http.StatusOK,
			results:         map[string][]byte{"https://google.com": []byte("google")},
		},

		{
			name:            "crawler_error",
			method:          http.MethodPost,
			body:            []byte(`["https://google.com"]`),
			urls:            []string{"https://google.com"},
			needCallCrawler: true,
			results:         nil,
			expectError:     errors.New("error"),
			expectedStatus:  http.StatusInternalServerError,
		},

		{
			name:            "crawler_error_cancel",
			method:          http.MethodPost,
			body:            []byte(`["https://google.com"]`),
			urls:            []string{"https://google.com"},
			needCallCrawler: true,
			results:         nil,
			expectError:     context.Canceled,
			expectedStatus:  http.StatusInternalServerError,
		},

		{
			name:            "invalid request",
			method:          http.MethodPost,
			body:            []byte(`[[`),
			needCallCrawler: false,
			urls:            []string{},
			expectedStatus:  http.StatusBadRequest,
		},

		{
			name:            "bad request",
			method:          http.MethodPost,
			body:            []byte(`["https://google.com","https://yandex.ru"]`),
			needCallCrawler: false,
			expectedStatus:  http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			mockCrawler := mocks.NewService(t)
			h := handlers.NewHTTPHandler(mockCrawler, 1, logger)

			if tc.needCallCrawler {
				mockCrawler.On("Crawl", ctx, tc.urls).
					Return(tc.results, tc.expectError).
					Once()
			}

			req := httptest.NewRequest(tc.method, "/", bytes.NewReader(tc.body))
			w := httptest.NewRecorder()
			h.Crawl(w, req)
			resp := w.Result()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			if resp.StatusCode == http.StatusOK {
				b, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var results map[string]string
				err = json.Unmarshal(b, &results)
				require.NoError(t, err)
				require.Len(t, results, len(tc.results))
				for r := range results {
					require.Equal(t, string(tc.results[r]), results[r])
				}
			}
		})
	}
}
