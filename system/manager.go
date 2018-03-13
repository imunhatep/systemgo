package system

import (
	"context"
	"log"
	"fmt"
	"time"
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

func (m *manager) Run(wg *sync.WaitGroup, ctx context.Context) {
	if m.isRunning {
		log.Fatal("[M] already running")
	}

	m.isRunning = true
	log.Println("[M] starting services")

	for i := range *m.serviceList {
		service := (*m.serviceList)[i]

		wg.Add(1)
		go service.Run(wg, ctx, m.outPipe, m.errPipe)
	}

	finished := make(chan struct{})
	go func() {
		wg.Wait()
		close(finished)
	}()

	// close Manager wait-group, and wait for processes
	wg.Done()

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
}

func (m *manager) wait() {
	if !m.isRunning {
		log.Fatal("[M] not started")
	}

	log.Println("[M] waiting services to finish")

	counter := len(*m.serviceList)
	fullStop := false
	for !fullStop && counter > 0 {
		counter -= 1
		fullStop = true

		for _, service := range *m.serviceList {
			fullStop = fullStop && !service.IsRunning()

			if service.IsRunning() {
				log.Println("[M] waiting", service.Name)
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	m.isRunning = false
}