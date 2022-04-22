package main

import (
	"sync"
	"sync/atomic"
)

func NewManager(container *Container) *Manager {
	manager := &Manager{
		container: container,
		level:     make(chan *Level, 10000),
	}


	return manager
}

type Manager struct {
	sync.Mutex

	container *Container
	id        int64
	level     chan *Level
	workers   []*Worker
}

func (m *Manager) AddWorker() {
	worker := &Worker{
		container: m.container,
		id:        atomic.AddInt64(&m.id, 1),
		level:     m.level,
	}

	m.Lock()
	m.workers = append(m.workers, worker)
	m.Unlock()

	go worker.Run()
}

func (m *Manager) Compute(level *Level) {
	m.level <- level
}
