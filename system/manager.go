package system

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type manager struct {
	serviceList *[]Service

	outPipe chan string
	errPipe chan string

	isRunning bool
}

func NewServiceManager(services []Service) *manager {
	m := new(manager)
	m.serviceList = &services

	m.isRunning = false

	bufSize := len(*m.serviceList)
	log.Printf("[M] buffer size: %d", bufSize)

	m.outPipe = make(chan string, bufSize)
	m.errPipe = make(chan string, bufSize)

	return m
}

func (m *manager) Run(ctx context.Context) {
	if m.isRunning {
		log.Fatal("[M] already running")
	}

	m.isRunning = true
	log.Println("[M] starting services")

	var wg sync.WaitGroup
	for i := range *m.serviceList {
		service := (*m.serviceList)[i]

		wg.Add(1)
		go func(){
			service.Run(ctx, m.outPipe, m.errPipe)
			wg.Done()
		}()
	}

	m.wait(&wg)
}

func (m manager) wait(wg *sync.WaitGroup) {
	if !m.isRunning {
		log.Fatal("[M] not started")
	}

	finished := make(chan struct{})
	go func() {
		wg.Wait()
		close(finished)
	}()

	for {
		select {
		case out := <-m.outPipe:
			fmt.Println(out)
		case err := <-m.errPipe:
			fmt.Println(err)
		case <-finished:
			log.Println("[M] finished")
			return
		}
	}

	m.isRunning = false
}
