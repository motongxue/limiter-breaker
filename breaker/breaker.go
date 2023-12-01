package breaker

import (
	"errors"
	"sync"
	"time"
)

const (
	STATE_CLOSE = iota
	STATE_OPEN
	STATE_HALF_OPEN
)

type Breaker struct {
	mu                sync.Mutex
	state             int
	failureThreshold  int
	successThreshold  int
	halfMaxRequest    int           // 半开半闭状态下，一个时间周期内最大请求数
	halfCycleReqCount int           // 半开半闭状态下，一个时间周期内请求数
	timeout           time.Duration // 超时时间
	failureCount      int           // 半开半闭状态下，请求失败次数多统计
	successCount      int           // 半开半闭状态下，请求成功次数多统计
	cycleStartTime    time.Time     // 周期时间
}

func NewBreaker(failureThreshold, successThreshold, halfMaxRequest int, timeout time.Duration) *Breaker {
	return &Breaker{
		state:            STATE_CLOSE, // 因为是断路器，所以初始为闭合状态
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		halfMaxRequest:   halfMaxRequest,
		timeout:          timeout,
	}
}

// Exec 参数的含义
func (b *Breaker) Exec(f func() error) error {
	b.before()
	if b.state == STATE_OPEN {
		return errors.New("断路器处于打开状态，无法访问服务")
	}
	if b.state == STATE_CLOSE {
		err := f()
		b.after(err)
		return err
	}
	if b.state == STATE_HALF_OPEN {
		if b.halfCycleReqCount < b.halfMaxRequest {
			err := f()
			b.after(err)
			return err
		} else {
			return errors.New("断路器处于半开闭状态，单位时间请求次数过多，请稍后再试")
		}
	}
	return nil
}
func (b *Breaker) before() {
	b.mu.Lock()
	defer b.mu.Unlock()
	// 检查Breaker的state
	switch b.state {
	case STATE_OPEN:
		if b.cycleStartTime.Add(b.timeout).Before(time.Now()) {
			b.state = STATE_HALF_OPEN
			b.reset()
			return
		}
	case STATE_HALF_OPEN:
		if b.successCount >= b.successThreshold {
			b.state = STATE_CLOSE
			b.reset()
			return
		}
		// 假如需要10次连续的成功才恢复为close状态，但在半开半闭状态下，一个时间周期内只允许4次请求，所以需要可以累计多个周期来完成连续成功次数
		if b.cycleStartTime.Add(b.timeout).Before(time.Now()) {
			b.cycleStartTime = time.Now()
			b.halfCycleReqCount = 0
			return
		}
	case STATE_CLOSE:
		if b.cycleStartTime.Add(b.timeout).Before(time.Now()) {
			b.reset()
		}
	}
}
func (b *Breaker) after(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err == nil {
		b.onSuccess()
	} else {
		b.onFailure()
	}
}

func (b *Breaker) reset() {
	b.successCount = 0
	b.failureCount = 0
	b.halfCycleReqCount = 0
	b.cycleStartTime = time.Now()
}

func (b *Breaker) onSuccess() {
	b.failureCount = 0
	if b.state == STATE_HALF_OPEN {
		// 成功次数+1仅在半开半闭状态下有意义
		b.successCount++
		b.halfCycleReqCount++
		if b.successCount >= b.successThreshold {
			b.state = STATE_CLOSE
			b.reset()
		}
	}
}

func (b *Breaker) onFailure() {
	b.successCount = 0
	b.failureCount++
	if b.state == STATE_HALF_OPEN || (b.state == STATE_CLOSE && b.failureCount >= b.failureThreshold) {
		b.state = STATE_OPEN
		b.reset()
		return
	}
}
