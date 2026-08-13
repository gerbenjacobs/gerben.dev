package internal

import "time"

type Ratelimiter struct {
	C chan bool
}

func NewRateLimiter(size int, sleep time.Duration) *Ratelimiter {
	r := &Ratelimiter{C: make(chan bool, size)}
	go func() {
		// seed the bucket, not actually necessary but gives a fast start.
		for i := 0; i < size; i++ {
			r.C <- true
		}

		// run forever and replenish the bucket.
		for {
			time.Sleep(sleep)
			r.C <- true
		}
	}()
	return r
}
