package limiter

import (
	"log"
)

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
		log.Println("take")
		return true
	default:
		log.Println("error take")
		return false
	}
}

func (c *ChanLimiter) Release() {
	log.Println("release")
	<-c.ch
}
