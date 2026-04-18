package limiter

type ConcurrencyLimiter struct {
	tokens chan struct{}
}

func NewConcurrencyLimiter(limit int) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		tokens: make(chan struct{}, limit),
	}
}

// Get 嘗試獲取執行權，若滿載則立即回傳 false (Non-blocking)
func (l *ConcurrencyLimiter) Get() bool {
	select {
	case l.tokens <- struct{}{}:
		return true
	default:
		return false
	}
}

// Put 釋放執行權
func (l *ConcurrencyLimiter) Put() {
	<-l.tokens
}
