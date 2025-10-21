package advanced

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func DemoAdvancedPatterns() {
	fmt.Println("Pipeline Pattern:")
	demoPipeline()
	
	fmt.Println("\nFan-Out/Fan-In Pattern:")
	demoFanOutFanIn()
	
	fmt.Println("\nCircuit Breaker Pattern:")
	demoCircuitBreaker()
	
	fmt.Println("\nSemaphore Pattern:")
	demoSemaphore()
}

func demoPipeline() {
	gen := func(nums ...int) <-chan int {
		out := make(chan int)
		go func() {
			for _, n := range nums {
				out <- n
			}
			close(out)
		}()
		return out
	}
	
	sq := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			for n := range in {
				out <- n * n
			}
			close(out)
		}()
		return out
	}
	
	multiply := func(in <-chan int, factor int) <-chan int {
		out := make(chan int)
		go func() {
			for n := range in {
				out <- n * factor
			}
			close(out)
		}()
		return out
	}
	
	c := gen(1, 2, 3, 4, 5)
	out := multiply(sq(c), 2)
	
	for n := range out {
		fmt.Printf("  Результат: %d\n", n)
	}
}

func demoFanOutFanIn() {
	producer := func(ctx context.Context) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for i := 1; i <= 10; i++ {
				select {
				case out <- i:
				case <-ctx.Done():
					return
				}
			}
		}()
		return out
	}
	
	worker := func(ctx context.Context, id int, in <-chan int) <-chan string {
		out := make(chan string)
		go func() {
			defer close(out)
			for num := range in {
				select {
				case out <- fmt.Sprintf("Worker %d обработал %d", id, num):
				case <-ctx.Done():
					return
				}
			}
		}()
		return out
	}
	
	fanIn := func(ctx context.Context, channels ...<-chan string) <-chan string {
		var wg sync.WaitGroup
		out := make(chan string)
		
		output := func(c <-chan string) {
			defer wg.Done()
			for msg := range c {
				select {
				case out <- msg:
				case <-ctx.Done():
					return
				}
			}
		}
		
		wg.Add(len(channels))
		for _, c := range channels {
			go output(c)
		}
		
		go func() {
			wg.Wait()
			close(out)
		}()
		
		return out
	}
	
	ctx := context.Background()
	in := producer(ctx)
	
	workers := make([]<-chan string, 3)
	for i := 0; i < 3; i++ {
		workers[i] = worker(ctx, i+1, in)
	}
	
	for result := range fanIn(ctx, workers...) {
		fmt.Printf("  %s\n", result)
	}
}

func demoCircuitBreaker() {
	type CircuitBreaker struct {
		maxFailures int
		failures    int
		lastFail    time.Time
		timeout     time.Duration
		mu          sync.Mutex
	}
	
	cb := &CircuitBreaker{
		maxFailures: 3,
		timeout:     2 * time.Second,
	}
	
	call := func(shouldFail bool) error {
		cb.mu.Lock()
		defer cb.mu.Unlock()
		
		if cb.failures >= cb.maxFailures {
			if time.Since(cb.lastFail) < cb.timeout {
				return fmt.Errorf("circuit breaker открыт")
			}
			cb.failures = 0
		}
		
		if shouldFail {
			cb.failures++
			cb.lastFail = time.Now()
			return fmt.Errorf("операция провалилась")
		}
		
		cb.failures = 0
		return nil
	}
	
	for i := 0; i < 6; i++ {
		err := call(i < 3)
		if err != nil {
			fmt.Printf("  Попытка %d: %v\n", i+1, err)
		} else {
			fmt.Printf("  Попытка %d: успешно\n", i+1)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func demoSemaphore() {
	type Semaphore struct {
		sem chan struct{}
	}
	
	NewSemaphore := func(max int) *Semaphore {
		return &Semaphore{
			sem: make(chan struct{}, max),
		}
	}
	
	sem := NewSemaphore(3)
	
	acquire := func() {
		sem.sem <- struct{}{}
	}
	
	release := func() {
		<-sem.sem
	}
	
	var wg sync.WaitGroup
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			acquire()
			fmt.Printf("  Задача %d запущена (макс 3 одновременно)\n", id)
			time.Sleep(200 * time.Millisecond)
			fmt.Printf("  Задача %d завершена\n", id)
			release()
		}(i)
	}
	wg.Wait()
}

func DemoCache() {
	fmt.Println("\nДемонстрация кэширования:")
	
	type CacheItem struct {
		Value      interface{}
		Expiration time.Time
	}
	
	type Cache struct {
		mu    sync.RWMutex
		items map[string]CacheItem
	}
	
	cache := &Cache{
		items: make(map[string]CacheItem),
	}
	
	set := func(key string, value interface{}, ttl time.Duration) {
		cache.mu.Lock()
		defer cache.mu.Unlock()
		cache.items[key] = CacheItem{
			Value:      value,
			Expiration: time.Now().Add(ttl),
		}
		fmt.Printf("  Сохранено в кэш: %s = %v\n", key, value)
	}
	
	get := func(key string) (interface{}, bool) {
		cache.mu.RLock()
		defer cache.mu.RUnlock()
		item, found := cache.items[key]
		if !found {
			return nil, false
		}
		if time.Now().After(item.Expiration) {
			return nil, false
		}
		return item.Value, true
	}
	
	set("user:1", "Иван Петров", 2*time.Second)
	set("user:2", "Мария Сидорова", 1*time.Second)
	
	if val, found := get("user:1"); found {
		fmt.Printf("  Получено из кэша: user:1 = %v\n", val)
	}
	
	time.Sleep(1500 * time.Millisecond)
	
	if val, found := get("user:2"); found {
		fmt.Printf("  Получено из кэша: user:2 = %v\n", val)
	} else {
		fmt.Println("  user:2 истек (TTL прошел)")
	}
	
	if val, found := get("user:1"); found {
		fmt.Printf("  Получено из кэша: user:1 = %v\n", val)
	}
}

