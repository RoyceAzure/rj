package pool

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkBasicPool_GetPut(b *testing.B) {
	bufferSize := 256

	p := NewBasicPool[byte](bufferSize)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := p.Get()
			// 模擬使用
			*buf = append(*buf, 1)
			// 重設長度避免攜帶容量外的長度
			*buf = (*buf)[:0]
			p.Put(buf)
		}
	})

	p.Stop()
}

func BenchmarkChannelPool_GetPut(b *testing.B) {
	poolSize := 32
	bufferSize := 256

	p := NewChannelPool[byte](poolSize, bufferSize)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := p.Get()
			*buf = append(*buf, 1)
			*buf = (*buf)[:0]
			p.Put(buf)
		}
	})

	p.Stop()
}

func BenchmarkShardChannelPool_GetPut(b *testing.B) {
	shardSize := 4
	poolSize := 16
	bufferSize := 256

	p := NewShardChannelPool[byte](shardSize, poolSize, bufferSize)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := p.Get()
			*buf = append(*buf, 1)
			*buf = (*buf)[:0]
			p.Put(buf)
		}
	})

	p.Stop()
}

// 輔助函數：使用指定倍數的 goroutine 執行 benchmark
func runBenchmarkWithGoroutines(b *testing.B, goroutineMultiplier int, fn func()) {
	numGoroutines := runtime.GOMAXPROCS(0) * goroutineMultiplier
	var counter int64

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	b.ResetTimer()
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for {
				// 使用 atomic 確保總共執行 b.N 次
				n := atomic.AddInt64(&counter, 1)
				if n > int64(b.N) {
					return
				}
				fn()
			}
		}()
	}
	wg.Wait()
}

func BenchmarkBasicPool_GetPut_2xGoroutines(b *testing.B) {
	bufferSize := 8000

	p := NewBasicPool[byte](bufferSize)
	b.ReportAllocs()
	b.ResetTimer()

	runBenchmarkWithGoroutines(b, 1000, func() {
		buf := p.Get()
		*buf = append(*buf, 1)
		*buf = (*buf)[:0]
		p.Put(buf)
	})

	p.Stop()
}

func BenchmarkBasicPool_GetPut_10xGoroutines(b *testing.B) {
	bufferSize := 8000

	p := NewBasicPool[byte](bufferSize)
	b.ReportAllocs()
	b.ResetTimer()

	runBenchmarkWithGoroutines(b, 10000, func() {
		buf := p.Get()
		*buf = append(*buf, 1)
		*buf = (*buf)[:0]
		p.Put(buf)
	})

	p.Stop()
}

func BenchmarkChannelPool_GetPut_2xGoroutines(b *testing.B) {
	poolSize := 256
	bufferSize := 8000

	p := NewChannelPool[byte](poolSize, bufferSize)
	time.Sleep(100 * time.Millisecond) // wait for ChannelPool
	b.ReportAllocs()
	b.ResetTimer()

	runBenchmarkWithGoroutines(b, 1000, func() {
		buf := p.Get()
		*buf = append(*buf, 1)
		*buf = (*buf)[:0]
		p.Put(buf)
	})

	p.Stop()
}

func BenchmarkChannelPool_GetPut_10xGoroutines(b *testing.B) {
	poolSize := 1024
	bufferSize := 8000

	p := NewChannelPool[byte](poolSize, bufferSize)
	time.Sleep(10 * time.Millisecond) // wait for ChannelPool
	b.ReportAllocs()
	b.ResetTimer()

	runBenchmarkWithGoroutines(b, 10000, func() {
		buf := p.Get()
		*buf = append(*buf, 1)
		*buf = (*buf)[:0]
		p.Put(buf)
	})

	p.Stop()
}

func BenchmarkShardChannelPool_GetPut_2xGoroutines(b *testing.B) {
	shardSize := 4
	poolSize := 16
	bufferSize := 8000

	p := NewShardChannelPool[byte](shardSize, poolSize, bufferSize)
	time.Sleep(10 * time.Millisecond) // wait for ChannelPool
	b.ReportAllocs()
	b.ResetTimer()

	runBenchmarkWithGoroutines(b, 1000, func() {
		buf := p.Get()
		*buf = append(*buf, 1)
		*buf = (*buf)[:0]
		p.Put(buf)
	})

	p.Stop()
}

func BenchmarkShardChannelPool_GetPut_10xGoroutines(b *testing.B) {
	shardSize := 16
	poolSize := 128
	bufferSize := 8000

	p := NewShardChannelPool[byte](shardSize, poolSize, bufferSize)
	time.Sleep(100 * time.Millisecond) // wait for ChannelPool
	b.ReportAllocs()
	b.ResetTimer()

	runBenchmarkWithGoroutines(b, 10000, func() {
		buf := p.Get()
		*buf = append(*buf, 1)
		*buf = (*buf)[:0]
		p.Put(buf)
	})

	p.Stop()
}
