package system

import (
	"log"
	"reflect"
	"fmt"
	"time"
)

type Manager struct {
	taskList *[]Service

	OutPipe chan string
	ErrPipe chan string

	exit chan bool

	isRunning bool
	isStopped bool
}

func (m *Manager) Run(services []Service) {
	if m.isRunning || m.isStopped {
		log.Fatal("[M] PM already started")
	}

	m.isStopped = false
	m.isRunning = true

	m.taskList = &services

	bufSize := len(*m.taskList)
	log.Printf("[M] Buf size: %d", bufSize)

	m.exit = make(chan bool, bufSize)
	m.OutPipe = make(chan string, bufSize)
	m.ErrPipe = make(chan string, bufSize)

	log.Println("[M] Starting services")

	for i := range *m.taskList {
		task := (*m.taskList)[i]

		//log.Printf("[M] \nName: %s\nExec: %s\nParams: %s\nRestart: %d\n\n", task.Name, task.Exec, task.Params, task.Restart)
		go task.Run(m.exit, m.OutPipe, m.ErrPipe)
	}

	m.listenStd()
}

func (m *Manager) Stop() {
	if !m.isRunning {
		log.Fatal("[M] ProcessManager not started")
	}
	m.isStopped = true

	log.Println("[M] Sending stop signal to tasks")
	for range *m.taskList {
		m.exit <- true
	}

	log.Println("[M] Waiting tasks to exit")

	counter := len(*m.taskList)
	fullStop := false
	for !fullStop && counter > 0 {
		counter -= 1
		fullStop = true

		for _, task := range *m.taskList {
			fullStop = fullStop && !task.IsRunning()
			if task.IsRunning() {
				log.Println("Waiting", task.Name)
			}
		}

		time.Sleep(1 * time.Second)
	}

	m.isRunning = false
}

func (m Manager) waitFor(tasks []chan error) {
	cases := make([]reflect.SelectCase, len(tasks))
	for i, ch := range tasks {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}

	remaining := len(cases)
	for remaining > 0 {
		chosen, value, ok := reflect.Select(cases)
		if !ok {
			log.Printf("[M] Read from channel %#v and received %s\n", tasks[chosen], value)
			continue
		}

		// The chosen channel has been closed, so zero out the channel to disable the case
		cases[chosen].Chan = reflect.ValueOf(nil)
		remaining -= 1
	}
}

func (m Manager) listenStd() {
	for {
		select {
		case out := <-m.OutPipe:
			fmt.Println(out)
		case err := <-m.ErrPipe:
			fmt.Println(err)
		}
	}
}
