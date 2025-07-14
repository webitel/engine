package utils

import "sync"

type PoolTask interface {
	Execute()
}

type Pool struct {
	mu    sync.Mutex
	size  int
	tasks chan PoolTask
	kill  chan struct{}
	wg    sync.WaitGroup
}

func NewPool(workers int, queueCount int) *Pool {
	pool := &Pool{
		tasks: make(chan PoolTask, queueCount),
		kill:  make(chan struct{}),
	}
	pool.Resize(workers)
	return pool
}

func (p *Pool) worker() {
	defer p.wg.Done()
	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				return
			}
			task.Execute()
		case <-p.kill:
			return
		}
	}
}

func (p *Pool) Resize(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for p.size < n {
		p.size++
		p.wg.Add(1)
		go p.worker()
	}
	for p.size > n {
		p.size--
		p.kill <- struct{}{}
	}
}

func (p *Pool) Close() {
	close(p.tasks)
}

func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) Exec(task PoolTask) {
	p.tasks <- task
}

func (p *Pool) Jobs() chan PoolTask {
	return p.tasks
}
