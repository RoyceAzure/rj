package limiter

import (
	"runtime"
	"sync"
	"testing"
)

func TestNewConcurrencyLimiter_GetPut(t *testing.T) {
	l := NewConcurrencyLimiter(2)
	if !l.Get() {
		t.Fatal("first Get should succeed")
	}
	if !l.Get() {
		t.Fatal("second Get should succeed")
	}
	if l.Get() {
		t.Fatal("third Get should fail when at capacity")
	}
	l.Put()
	if !l.Get() {
		t.Fatal("Get should succeed after Put releases a slot")
	}
	if l.Get() {
		t.Fatal("should still be at capacity")
	}
	l.Put()
	l.Put()
}

func TestConcurrencyLimiter_ZeroLimit(t *testing.T) {
	l := NewConcurrencyLimiter(0)
	if l.Get() {
		t.Fatal("Get on zero-capacity limiter should not acquire (non-blocking)")
	}
}

func TestConcurrencyLimiter_ParallelLimitOne(t *testing.T) {
	l := NewConcurrencyLimiter(1)
	const workers = 128
	const iters = 64

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				for !l.Get() {
					runtime.Gosched()
				}
				l.Put()
			}
		}()
	}
	wg.Wait()
}

func TestConcurrencyLimiter_StressThenDrain(t *testing.T) {
	const limit = 8
	l := NewConcurrencyLimiter(limit)
	const workers = 64
	const iters = 50

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				for !l.Get() {
					runtime.Gosched()
				}
				l.Put()
			}
		}()
	}
	wg.Wait()

	for i := 0; i < limit; i++ {
		if !l.Get() {
			t.Fatalf("expected empty limiter after stress, Get %d should succeed", i)
		}
	}
	if l.Get() {
		t.Fatal("limiter should be at capacity after refill")
	}
	for i := 0; i < limit; i++ {
		l.Put()
	}
}

func TestConcurrencyLimiter_PutAfterFullCycle(t *testing.T) {
	l := NewConcurrencyLimiter(1)
	if !l.Get() {
		t.Fatal("Get should succeed")
	}
	l.Put()
	if !l.Get() {
		t.Fatal("Get after full Put should succeed")
	}
	l.Put()
}
