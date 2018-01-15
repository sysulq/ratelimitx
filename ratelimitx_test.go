package ratelimitx

import (
	"testing"
	"time"

	"golang.org/x/time/rate"

	"github.com/stretchr/testify/assert"
)

// case：达到预定阀值时，allow应返回false
func TestAllow(t *testing.T) {
	l := New("localhost:11211")
	l.Reset("test_id", time.Minute)

	rate, delay, allow := l.Allow("test_id", 1, time.Minute)
	assert.True(t, allow)
	assert.Equal(t, uint64(1), rate)
	assert.Condition(t, func() bool { return delay <= time.Minute })

	rate, _, allow = l.Allow("test_id", 1, time.Minute)
	assert.False(t, allow)
	assert.Equal(t, uint64(2), rate)
}

// case：先达到预定阀值，然后等待过期，此时再次请求应顺利通过
func TestExpire(t *testing.T) {
	l := New("localhost:11211")
	l.Reset("test_id", time.Second)

	rate, delay, allow := l.Allow("test_id", 1, time.Second)
	assert.True(t, allow)
	assert.Equal(t, uint64(1), rate)
	assert.Condition(t, func() bool { return delay <= time.Second })

	rate, _, allow = l.Allow("test_id", 1, time.Second)
	assert.False(t, allow)
	assert.Equal(t, uint64(2), rate)

	time.Sleep(time.Second)

	rate, delay, allow = l.Allow("test_id", 1, time.Second)
	assert.True(t, allow)
	assert.Equal(t, uint64(1), rate)
	assert.Condition(t, func() bool { return delay <= time.Second })
}

// case：测试AllowMinute
func TestAllowMinute(t *testing.T) {
	l := New("localhost:11211")
	l.Reset("test_id", time.Minute)

	rate, delay, allow := l.AllowMinute("test_id", 1)
	assert.True(t, allow)
	assert.Equal(t, uint64(1), rate)
	assert.Condition(t, func() bool { return delay <= time.Minute })

	rate, _, allow = l.Allow("test_id", 1, time.Minute)
	assert.False(t, allow)
	assert.Equal(t, uint64(2), rate)
}

// case：测试AllowHour
func TestAllowHour(t *testing.T) {
	l := New("localhost:11211")
	l.Reset("test_id", time.Hour)

	rate, delay, allow := l.AllowHour("test_id", 1)
	assert.True(t, allow)
	assert.Equal(t, uint64(1), rate)
	assert.Condition(t, func() bool { return delay <= time.Hour })

	rate, _, allow = l.Allow("test_id", 1, time.Hour)
	assert.False(t, allow)
	assert.Equal(t, uint64(2), rate)
}

// case：测试AllowRate
func TestAllowRate(t *testing.T) {
	l := New("localhost:11211")
	l.ResetRate("test_id", rate.Every(time.Second))

	delay, allow := l.AllowRate("test_id", rate.Every(time.Second))
	assert.True(t, allow)
	assert.Condition(t, func() bool { return delay <= time.Second })

	delay, allow = l.AllowRate("test_id", rate.Every(time.Second))
	assert.False(t, allow)
	assert.Condition(t, func() bool { return delay <= time.Second })
}

// case：如memcache挂了，且没定义fallback，则直接失败（有待讨论）
func TestMemcacheUnavailable(t *testing.T) {
	l := New("localhost:11211")
	l.Reset("test_id", time.Second)

	count, delay, allow := l.Allow("TestMemcacheUnavailable", 1, time.Second)
	assert.False(t, allow)
	assert.Equal(t, uint64(0), count)
	assert.Condition(t, func() bool { return delay <= time.Second })

	count, delay, allow = l.Allow("TestMemcacheUnavailable", 1, time.Second)
	assert.False(t, allow)
	assert.Equal(t, uint64(0), count)
	assert.Condition(t, func() bool { return delay <= time.Second })
}

// case：如memcache挂了，可fallback到单机限速
func TestMemcacheUnavailableWithFallback(t *testing.T) {
	l := New("localhost:11211")
	l.Reset("test_id", time.Second)
	l.Fallback = rate.NewLimiter(rate.Every(time.Second), 1)

	count, delay, allow := l.Allow("TestMemcacheUnavailableWithFallback", 1, time.Second)
	assert.True(t, allow)
	assert.Equal(t, uint64(1), count)
	assert.Condition(t, func() bool { return delay <= time.Second })

	count, delay, allow = l.Allow("TestMemcacheUnavailableWithFallback", 1, time.Second)
	assert.False(t, allow)
	assert.Equal(t, uint64(1), count)
	assert.Condition(t, func() bool { return delay <= time.Second })

	count, delay, allow = l.Allow("TestMemcacheUnavailableWithFallback", 1, time.Second)
	assert.False(t, allow)
	assert.Equal(t, uint64(1), count)
	assert.Condition(t, func() bool { return delay <= time.Second })
}

// benchmark AllowSecond
func BenchmarkAllowSecond(b *testing.B) {
	l := New("localhost:11211")
	l.Reset("test_id", time.Second)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Allow("test_id", 1, time.Second)
	}
}

// benchmark AllowMinute
func BenchmarkAllowMinute(b *testing.B) {
	l := New("localhost:11211")
	l.Reset("test_id", time.Minute)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.AllowMinute("test_id", 1)
	}
}

// benchmark AllowHour
func BenchmarkAllowHour(b *testing.B) {
	l := New("localhost:11211")
	l.Reset("test_id", time.Hour)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.AllowHour("test_id", 1)
	}
}

// benchmark memcache unavailable with fallback
func BenchmarkMemcacheUnavailableWithFallback(b *testing.B) {
	l := New("localhost:11211")
	l.Reset("test_id", time.Hour)
	l.Fallback = rate.NewLimiter(rate.Every(time.Second), 1)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.AllowHour("test_id", 1)
	}
}

// benchmark memcache unavailable without fallback
func BenchmarkMemcacheUnavailableWithoutFallback(b *testing.B) {
	l := New("localhost:11211")
	l.Reset("test_id", time.Hour)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.AllowHour("test_id", 1)
	}
}
