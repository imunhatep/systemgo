package system

import (
	"context"
	"log"
	"fmt"
	"time"
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
	log.Printf("[M] Buf size: %d", bufSize)

	m.outPipe = make(chan string, bufSize)
	m.errPipe = make(chan string, bufSize)

	return m
}

func (m *manager) Run(ctx context.Context) {
	if m.isRunning {
		log.Fatal("[M] PM already started")
	}

	m.isRunning = true

	log.Println("[M] Starting services")
	for i := range *m.serviceList {
		service := (*m.serviceList)[i]

		//log.Printf("[M] \nName: %s\nExec: %s\nParams: %s\nRestart: %d\n\n", service.Name, service.Exec, service.Params, service.Restart)
		go service.Run(ctx, m.outPipe, m.errPipe)
	}

	for {
		select {
		case out := <-m.outPipe:
			fmt.Println(out)
		case err := <-m.errPipe:
			fmt.Println(err)
		case <-ctx.Done():
			m.stop()
			return
		}
	}
}

func (m *manager) stop() {
	if !m.isRunning {
		log.Fatal("[M] ProcessManager not started")
	}

	log.Println("[M] Waiting services to finish")

	counter := len(*m.serviceList)
	fullStop := false
	for !fullStop && counter > 0 {
		counter -= 1
		fullStop = true

		for _, service := range *m.serviceList {
			fullStop = fullStop && !service.IsRunning()
			if service.IsRunning() {
				log.Println("Waiting", service.Name)
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	m.isRunning = false
}