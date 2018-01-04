package ratelimitx

import (
	"fmt"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/hnlq715/gobreak"
	"golang.org/x/time/rate"
)

const prefix = "ratelimitx"

// Limiter controls how frequently events are allowed to happen.
type Limiter struct {
	client *memcache.Client

	// Optional fallback limiter used when Memcache is unavailable.
	Fallback *rate.Limiter
}

// New creates a Limiter
func New(server ...string) *Limiter {
	return &Limiter{
		client: memcache.New(server...),
	}
}

// NewWithMemcache creats a Limiter
func NewWithMemcache(client *memcache.Client) *Limiter {
	return &Limiter{
		client: client,
	}
}

// Reset resets the rate limit for the name in the given rate limit window.
func (l *Limiter) Reset(name string, dur time.Duration) error {
	udur := int64(dur / time.Second)
	slot := time.Now().Unix() / udur

	name = allowName(name, slot)
	return l.client.Delete(name)
}

// ResetRate resets the rate limit for the name and limit.
func (l *Limiter) ResetRate(name string, rateLimit rate.Limit) error {
	if rateLimit == 0 {
		return nil
	}
	if rateLimit == rate.Inf {
		return nil
	}

	dur := time.Second
	limit := int64(rateLimit)
	if limit == 0 {
		limit = 1
		dur *= time.Duration(1 / rateLimit)
	}
	slot := time.Now().UnixNano() / dur.Nanoseconds()

	name = allowRateName(name, dur, slot)
	return l.client.Delete(name)
}

// AllowN reports whether an event with given name may happen at time now.
// It allows up to maxn events within duration dur, with each interaction
// incrementing the limit by n.
func (l *Limiter) AllowN(name string, maxn uint64, dur time.Duration, n int64) (count uint64, delay time.Duration, allow bool) {

	udur := int64(dur / time.Second)
	utime := time.Now().Unix()
	slot := utime / udur
	delay = time.Duration((slot+1)*udur-utime) * time.Second

	name = allowName(name, slot)
	count, err := l.incr(name, dur, n)
	if err == nil {
		allow = count <= maxn
	} else {
		if l.Fallback != nil {
			allow = l.Fallback.Allow()
			count = uint64(l.Fallback.Limit())
		}
	}

	return count, delay, allow
}

// Allow is shorthand for AllowN(name, max, dur, 1).
func (l *Limiter) Allow(name string, maxn uint64, dur time.Duration) (count uint64, delay time.Duration, allow bool) {
	return l.AllowN(name, maxn, dur, 1)
}

// AllowMinute is shorthand for Allow(name, maxn, time.Minute).
func (l *Limiter) AllowMinute(name string, maxn uint64) (count uint64, delay time.Duration, allow bool) {
	return l.Allow(name, maxn, time.Minute)
}

// AllowHour is shorthand for Allow(name, maxn, time.Hour).
func (l *Limiter) AllowHour(name string, maxn uint64) (count uint64, delay time.Duration, allow bool) {
	return l.Allow(name, maxn, time.Hour)
}

// AllowRate reports whether an event may happen at time now.
// It allows up to rateLimit events each second.
func (l *Limiter) AllowRate(name string, rateLimit rate.Limit) (delay time.Duration, allow bool) {
	if rateLimit == 0 {
		return 0, false
	}
	if rateLimit == rate.Inf {
		return 0, true
	}

	dur := time.Second
	limit := uint64(rateLimit)
	if limit == 0 {
		limit = 1
		dur *= time.Duration(1 / rateLimit)
	}
	now := time.Now()
	slot := now.UnixNano() / dur.Nanoseconds()

	if l.Fallback != nil {
		allow = l.Fallback.Allow()
	}

	name = allowRateName(name, dur, slot)
	count, err := l.incr(name, dur, 1)
	if err == nil {
		allow = count <= limit
	}

	if !allow {
		delay = time.Duration(slot+1)*dur - time.Duration(now.UnixNano())
	}

	return delay, allow
}

func (l *Limiter) incr(name string, dur time.Duration, n int64) (incr uint64, err error) {
	// 快速失败，避免因memcache挂掉导致影响业务
	err = gobreak.Do("memcache incr", func() error {
		// 先Incr
		incr, err = l.client.Increment(name, uint64(n))
		if err == memcache.ErrCacheMiss {
			// 若返回ErrCacheMiss则表明Memcache中没有该键值
			// 需手动Set并初始化其键值，过期时间根据dur计算
			incr = 1
			err = l.client.Set(&memcache.Item{
				Key:        name,
				Value:      []byte("1"),
				Expiration: int32(dur / time.Second),
			})
		}
		return err
	}, nil)

	return incr, err
}

func allowName(name string, slot int64) string {
	return fmt.Sprintf("%s:%s-%d", prefix, name, slot)
}

func allowRateName(name string, dur time.Duration, slot int64) string {
	return fmt.Sprintf("%s:%s-%d-%d", prefix, name, dur, slot)
}
