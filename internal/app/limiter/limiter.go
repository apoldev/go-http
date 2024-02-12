package limiter

type ChanLimiter struct {
	ch chan struct{}
}

func NewChanLimiter(limit int) *ChanLimiter {
	return &ChanLimiter{
		ch: make(chan struct{}, limit),
	}
}

func (c *ChanLimiter) Take() bool {
	select {
	case c.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

func (c *ChanLimiter) Release() {
	<-c.ch
}
