package limiter_test

import (
	"github.com/apoldev/go-http/internal/app/limiter"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

type limiterI interface {
	Take() bool
	Release()
}

func TestLimiter(t *testing.T) {

	cases := []struct {
		name               string
		limiter            limiterI
		expectedStatusesOK int
		count              int
	}{
		{
			name:               "chan_limiter_5",
			limiter:            limiter.NewChanLimiter(5),
			count:              10,
			expectedStatusesOK: 5,
		},

		{
			name:               "chan_limiter_1",
			limiter:            limiter.NewChanLimiter(1),
			count:              5,
			expectedStatusesOK: 1,
		},

		{
			name:               "atom_limiter_5",
			limiter:            limiter.NewAtomLimiter(5),
			count:              10,
			expectedStatusesOK: 5,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			results := make([]int, c.count)

			wg := sync.WaitGroup{}
			wg.Add(c.count)
			for i := 0; i < c.count; i++ {
				go func(i int) {
					defer wg.Done()
					if !c.limiter.Take() {
						return
					}
					defer c.limiter.Release()
					results[i] = 1

					time.Sleep(time.Millisecond * 100)
				}(i)
			}
			wg.Wait()

			var count int
			for j := range results {
				count += results[j]
			}

			require.Equal(t, c.expectedStatusesOK, count)

		})
	}

}

func BenchmarkChanAtom(b *testing.B) {

	cases := []struct {
		name    string
		limiter limiterI
		sleep   time.Duration
	}{
		{
			name:    "chan_1000",
			limiter: limiter.NewChanLimiter(1000),
		},
		{
			name:    "atom_1000",
			limiter: limiter.NewAtomLimiter(1000),
		},
	}

	for _, c := range cases {
		c := c
		b.Run(c.name, func(b *testing.B) {
			wg := sync.WaitGroup{}
			for i := 0; i < b.N; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if c.limiter.Take() {
						defer c.limiter.Release()
					}
				}()
				wg.Wait()
			}
		})
	}

}
