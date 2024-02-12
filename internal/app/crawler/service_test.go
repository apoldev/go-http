package crawler_test

import (
	"bytes"
	"context"
	"github.com/apoldev/go-http/internal/app/crawler"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	time.Sleep(100 * time.Millisecond)
	select {
	case <-req.Context().Done():
		return nil, req.Context().Err()
	default:
		return f(req), nil
	}
}

func getFakeHttpClient(datas map[string][]byte) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(func(req *http.Request) *http.Response {
			for url, data := range datas {
				if req.URL.String() == url {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(data)),
					}
				}
			}
			return nil
		}),
	}
}

func TestService_Crawl(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Fake external servers
	urls := map[string][]byte{
		"https://tracktrace.dpd.com.pl/EN/findPackage": []byte(`[1,2,3]`),
		"http://google.com":                            []byte(`[4,5,6]`),
		"http://yandex.ru":                             []byte(`<html><body>hello</body></html>`),
		"http://mail.ru":                               []byte(`<html><body>mail</body></html>`),
	}
	client := getFakeHttpClient(urls)

	ctx1 := context.Background()
	ctx2, cancel := context.WithCancel(context.Background())
	cancel()

	cases := []struct {
		name                    string
		workerCount             int
		crawlerRequestTimeoutMs int
		urls                    []string
		expectedData            map[string][]byte
		expectedError           error
		ctx                     context.Context
	}{
		{
			name:                    "valid_data",
			workerCount:             1,
			crawlerRequestTimeoutMs: 1000,
			urls:                    []string{"https://tracktrace.dpd.com.pl/EN/findPackage"},
			expectedData:            map[string][]byte{"https://tracktrace.dpd.com.pl/EN/findPackage": []byte(`[1,2,3]`)},
			ctx:                     ctx1,
		},

		{
			name:                    "valid_data",
			workerCount:             1,
			crawlerRequestTimeoutMs: 1000,
			urls:                    []string{"https://tracktrace.dpd.com.pl/EN/findPackage"},
			expectedData:            map[string][]byte{"https://tracktrace.dpd.com.pl/EN/findPackage": []byte(`[1,2,3]`)},
			ctx:                     ctx1,
		},

		{
			name:                    "cancel_client",
			workerCount:             1,
			crawlerRequestTimeoutMs: 1000,
			urls:                    []string{"http://google.com", "http://yandex.ru", "http://mail.ru"},
			expectedData:            map[string][]byte{},
			ctx:                     ctx2,
		},

		{
			name:                    "empty_urls",
			workerCount:             1,
			crawlerRequestTimeoutMs: 1000,
			urls:                    []string{},
			expectedData:            map[string][]byte{},
			ctx:                     ctx1,
		},

		{
			name:                    "cancel_by_timeout",
			workerCount:             1,
			crawlerRequestTimeoutMs: 50,
			urls:                    []string{"http://google.com", "http://yandex.ru", "http://mail.ru"},
			expectedData:            map[string][]byte{},
			ctx:                     ctx1,
			expectedError:           context.DeadlineExceeded,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//t.Parallel()
			c := crawler.New(tc.workerCount, tc.crawlerRequestTimeoutMs, client, logger)
			data, err := c.Crawl(tc.ctx, tc.urls)

			if tc.expectedError != nil {
				require.ErrorContains(t, err, tc.expectedError.Error())
				return
			}

			require.Len(t, data, len(tc.expectedData))
			for i := range tc.expectedData {
				if v, ok := data[i]; ok {
					require.Equal(t, tc.expectedData[i], v)
				}
			}

		})
	}

}
