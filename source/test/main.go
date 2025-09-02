package main



import (
	"fmt"
	"sync"
	"time"
)

// 工人函数
func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		fmt.Printf("Worker %d 开始任务 %d\n", id, job)
		time.Sleep(500 * time.Millisecond) // 模拟耗时
		fmt.Printf("Worker %d 完成任务 %d\n", id, job)
	}
}

func main() {
	const workerCount = 3
	jobs := make(chan int, 10)
	var wg sync.WaitGroup

	// 启动 worker
	for w := 1; w <= workerCount; w++ {
		wg.Add(1)
		wg.Go(
			func() {
				worker(w, jobs, &wg)
			})
	}

	// 投递任务
	for j := 1; j <= 100; j++ {
		jobs <- j
	}
	close(jobs)

	// 等待 worker 全部退出
	wg.Wait()
}
