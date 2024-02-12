package limiter

import "sync/atomic"

type AtomLimiter struct {
	limit int32
}

func NewAtomLimiter(limit int) *AtomLimiter {
	return &AtomLimiter{
		limit: int32(limit),
	}
}

func (c *AtomLimiter) Take() bool {
	if atomic.LoadInt32(&c.limit) <= 0 {
		return false
	}
	atomic.AddInt32(&c.limit, -1)
	return true
}

func (c *AtomLimiter) Release() {
	atomic.AddInt32(&c.limit, 1)
}
