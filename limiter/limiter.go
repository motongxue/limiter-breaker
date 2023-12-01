package limiter

import (
	"sync"
	"time"
)

type Limiter struct {
	tb *TokenBucket
}

type TokenBucket struct {
	// 锁
	mu sync.Mutex
	// 桶大小
	size int
	// 当前token数
	count int
	// 填充速率，即每隔多久补充一个token
	rateLimit time.Duration
	//最后成功请求的时间
	lastRequestTime time.Time
}

func NewLimiter(r time.Duration, size int) *Limiter {
	return &Limiter{
		tb: &TokenBucket{
			rateLimit: r,
			size:      size,
			count:     size,
		},
	}
}

func (tb *TokenBucket) fillToken() {
	tb.count += tb.getFillTokenCount()
}

func (tb *TokenBucket) getFillTokenCount() int {
	if tb.count < tb.size && !tb.lastRequestTime.IsZero() {
		// IsZero代表是否为零值，即是否为第一次请求
		duration := time.Now().Sub(tb.lastRequestTime)
		count := int(duration / tb.rateLimit)
		if tb.size-tb.count >= count {
			return count
		} else {
			return tb.size - tb.count
		}
	}
	return 0
}

func (l *Limiter) Allow() bool {
	// 填充
	l.tb.fillToken()
	if l.tb.count > 0 {
		l.tb.count--
		l.tb.lastRequestTime = time.Now()
		return true
	}
	return false
}
