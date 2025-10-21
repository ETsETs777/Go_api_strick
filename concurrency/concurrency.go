package concurrency

import (
	"fmt"
	"sync"
	"time"
)

func DemoGoroutines() {
	fmt.Println("Запуск горутин...")
	
	go printNumbers("Горутина 1")
	go printNumbers("Горутина 2")
	
	go func() {
		for i := 1; i <= 3; i++ {
			fmt.Printf("  Анонимная горутина: %d\n", i)
			time.Sleep(150 * time.Millisecond)
		}
	}()
	
	time.Sleep(1 * time.Second)
	
	var wg sync.WaitGroup
	
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("  Worker %d начал работу\n", id)
			time.Sleep(200 * time.Millisecond)
			fmt.Printf("  Worker %d завершил работу\n", id)
		}(i)
	}
	
	wg.Wait()
	fmt.Println("Все workers завершены")
}

func printNumbers(name string) {
	for i := 1; i <= 3; i++ {
		fmt.Printf("  %s: %d\n", name, i)
		time.Sleep(100 * time.Millisecond)
	}
}

func DemoChannels() {
	ch := make(chan int)
	
	go func() {
		for i := 1; i <= 5; i++ {
			ch <- i
			fmt.Printf("  Отправлено: %d\n", i)
		}
		close(ch)
	}()
	
	for num := range ch {
		fmt.Printf("  Получено: %d\n", num)
		time.Sleep(100 * time.Millisecond)
	}
	
	bufferedCh := make(chan string, 3)
	bufferedCh <- "первое"
	bufferedCh <- "второе"
	bufferedCh <- "третье"
	
	fmt.Printf("Буферизованный канал: %s, %s, %s\n", 
		<-bufferedCh, <-bufferedCh, <-bufferedCh)
	
	messages := make(chan string)
	go sendMessages(messages)
	receiveMessages(messages)
}

func sendMessages(ch chan<- string) {
	messages := []string{"Hello", "Go", "Channels"}
	for _, msg := range messages {
		ch <- msg
		time.Sleep(100 * time.Millisecond)
	}
	close(ch)
}

func receiveMessages(ch <-chan string) {
	for msg := range ch {
		fmt.Printf("  Сообщение: %s\n", msg)
	}
}

func DemoSelect() {
	ch1 := make(chan string)
	ch2 := make(chan string)
	
	go func() {
		time.Sleep(200 * time.Millisecond)
		ch1 <- "из канала 1"
	}()
	
	go func() {
		time.Sleep(100 * time.Millisecond)
		ch2 <- "из канала 2"
	}()
	
	for i := 0; i < 2; i++ {
		select {
		case msg1 := <-ch1:
			fmt.Printf("  Получено %s\n", msg1)
		case msg2 := <-ch2:
			fmt.Printf("  Получено %s\n", msg2)
		case <-time.After(300 * time.Millisecond):
			fmt.Println("  Таймаут!")
		}
	}
	
	select {
	case msg := <-ch1:
		fmt.Println(msg)
	default:
		fmt.Println("  Нет доступных сообщений (default)")
	}
}

func DemoWorkerPool() {
	const numWorkers = 3
	const numJobs = 10
	
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)
	
	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}
	
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)
	
	go func() {
		wg.Wait()
		close(results)
	}()
	
	fmt.Println("Результаты:")
	for result := range results {
		fmt.Printf("  Результат: %d\n", result)
	}
}

func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		fmt.Printf("  Worker %d обрабатывает задачу %d\n", id, job)
		time.Sleep(100 * time.Millisecond)
		results <- job * 2
	}
}

func DemoMutex() {
	counter := SafeCounter{value: 0}
	var wg sync.WaitGroup
	
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Inc()
		}()
	}
	
	wg.Wait()
	fmt.Printf("Финальное значение счетчика: %d\n", counter.Value())
}

type SafeCounter struct {
	mu    sync.Mutex
	value int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}
