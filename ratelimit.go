// Copyright (c) 2015, Bruce Marriner
// All rights reserved.
// some modifications for PoC

package ddp

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// customRateLimit holds information for defining a custom rate limit
type customRateLimit struct {
	suffix   string
	requests int
	reset    time.Duration
}

// RateLimiter holds all ratelimit buckets
type RateLimiter struct {
	sync.Mutex
	global           *int64
	buckets          map[string]*Bucket // map endpoints to bucket
	globalRateLimit  time.Duration
	customRateLimits []*customRateLimit
}

// NewRateLimiter returns a new RateLimiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*Bucket),
		global:  new(int64),
	}
}

// UncleDrew gets bucket
func (r *RateLimiter) UncleDrew(key string) *Bucket {
	r.Lock()
	defer r.Unlock()

	if bucket, ok := r.buckets[key]; ok {
		return bucket
	}

	b := &Bucket{
		Remaining: 1,
		Key:       key,
		global:    r.global,
	}

	// Check if there is a custom ratelimit set for this bucket ID.
	for _, rl := range r.customRateLimits {
		if strings.HasSuffix(b.Key, rl.suffix) {
			b.customRateLimit = rl
			break
		}
	}

	r.buckets[key] = b
	return b
}

// GetWaitTime returns the duration to wait for a Bucket
func (r *RateLimiter) GetWaitTime(b *Bucket, minRemaining int) time.Duration {
	if b.Remaining < minRemaining && b.reset.After(time.Now()) {
		return b.reset.Sub(time.Now())
	}

	// Check for global ratelimits
	sleepTo := time.Unix(0, atomic.LoadInt64(r.global))
	if now := time.Now(); now.Before(sleepTo) {
		return sleepTo.Sub(now)
	}

	return 0
}

// LockBucket locks until a request can be made
func (r *RateLimiter) LockBucket(key string) *Bucket {
	return r.LockBucketObject(r.UncleDrew(key))
}

// TODO edit LockBucketObject to use channel instead of sleep?
// LockBucketObject locks an already resolved bucket until a request can be made
func (r *RateLimiter) LockBucketObject(b *Bucket) *Bucket {
	b.Lock()

	if wait := r.GetWaitTime(b, 1); wait > 0 {
		time.Sleep(wait)
	}

	b.Remaining--
	return b
}

// Bucket represents a ratelimit bucket, each bucket gets ratelimited individually (-global ratelimits)
type Bucket struct {
	sync.Mutex
	Key       string
	Remaining int
	limit     int
	reset     time.Time
	global    *int64

	lastReset       time.Time
	customRateLimit *customRateLimit
	Userdata        interface{}
}

// Release unlocks the bucket and reads the headers to update the buckets ratelimit info
// and locks up the whole thing in case if there's a global ratelimit.
func (b *Bucket) Release(headers http.Header) error {
	defer b.Unlock()

	// Check if the bucket uses a custom ratelimiter
	if rl := b.customRateLimit; rl != nil {
		if time.Now().Sub(b.lastReset) >= rl.reset {
			b.Remaining = rl.requests - 1
			b.lastReset = time.Now()
		}
		if b.Remaining < 1 {
			b.reset = time.Now().Add(rl.reset)
		}
		return nil
	}

	if headers == nil {
		return nil
	}

	remaining := headers.Get("X-RateLimit-Remaining")
	reset := headers.Get("X-RateLimit-Reset")
	global := headers.Get("X-RateLimit-Global")
	retryAfter := headers.Get("Retry-After")

	// Update global and per bucket reset time if the proper headers are available
	// If global is set, then it will block all buckets until after Retry-After
	// If Retry-After without global is provided it will use that for the new reset
	// time since it's more accurate than X-RateLimit-Reset.
	// If Retry-After after is not proided, it will update the reset time from X-RateLimit-Reset
	if retryAfter != "" {
		parsedAfter, err := strconv.ParseInt(retryAfter, 10, 64)
		if err != nil {
			return err
		}

		resetAt := time.Now().Add(time.Duration(parsedAfter) * time.Millisecond)

		// Lock either this single bucket or all buckets
		if global != "" {
			atomic.StoreInt64(b.global, resetAt.UnixNano())
		} else {
			b.reset = resetAt
		}
	} else if reset != "" {
		// Calculate the reset time by using the date header returned from discord
		discordTime, err := http.ParseTime(headers.Get("Date"))
		if err != nil {
			return err
		}

		unix, err := strconv.ParseInt(reset, 10, 64)
		if err != nil {
			return err
		}

		// Calculate the time until reset and add it to the current local time
		// some extra time is added because without it i still encountered 429's.
		// The added amount is the lowest amount that gave no 429's
		// in 1k requests
		delta := time.Unix(unix, 0).Sub(discordTime) + time.Millisecond*250
		b.reset = time.Now().Add(delta)
	}

	// Update remaining if header is present
	if remaining != "" {
		parsedRemaining, err := strconv.ParseInt(remaining, 10, 32)
		if err != nil {
			return err
		}
		b.Remaining = int(parsedRemaining)
	}

	return nil
}
