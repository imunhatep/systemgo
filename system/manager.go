package system

import (
	"log"
	"reflect"
	"fmt"
	"time"
)

type Manager struct {
	taskList *[]Task

	MemPipe chan string
	OutPipe chan string
	ErrPipe chan string

	exit chan bool

	isRunning bool
	isStopped bool
}

func (t *Manager) Run(tasks []Task) {
	if t.isRunning || t.isStopped {
		log.Fatal("[MANAGER] PM already started")
	}

	t.isStopped = false
	t.isRunning = true

	t.taskList = &tasks

	bufSize := len(*t.taskList)
	log.Printf("[MANAGER] Buf size: %d", bufSize)

	t.exit = make(chan bool, bufSize)
	t.MemPipe = make(chan string, bufSize)
	t.OutPipe = make(chan string, bufSize)
	t.ErrPipe = make(chan string, bufSize)

	log.Println("[MANAGER] Starting TASKS")

	for i := range *t.taskList {
		task := (*t.taskList)[i]

		//log.Printf("[MANAGER] \nName: %s\nExec: %s\nParams: %s\nRestart: %d\n\n", task.Name, task.Exec, task.Params, task.Restart)
		go task.Run(t.exit, t.OutPipe, t.ErrPipe)
	}

	t.listenStd()
}

func (t *Manager) Stop() {
	if !t.isRunning {
		log.Fatal("[MANAGER] ProcessManager not started")
	}
	t.isStopped = true

	log.Println("[MANAGER] Sending stop signal to tasks")
	for range *t.taskList {
		t.exit <- true
	}

	log.Println("[MANAGER] Waiting tasks to exit")

	counter := len(*t.taskList)
	fullStop := false
	for !fullStop && counter > 0 {
		counter -= 1
		fullStop = true

		for _, task := range *t.taskList {
			fullStop = fullStop && !task.IsRunning()
			if task.IsRunning() {
				log.Println("Waiting", task.Name)
			}
		}

		time.Sleep(1 * time.Second)
	}

	t.isRunning = false
}

func (t Manager) waitFor(tasks []chan error) {
	cases := make([]reflect.SelectCase, len(tasks))
	for i, ch := range tasks {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}

	remaining := len(cases)
	for remaining > 0 {
		chosen, value, ok := reflect.Select(cases)
		if !ok {
			log.Printf("[MANAGER] Read from channel %#v and received %s\n", tasks[chosen], value)
			continue
		}

		// The chosen channel has been closed, so zero out the channel to disable the case
		cases[chosen].Chan = reflect.ValueOf(nil)
		remaining -= 1
	}
}

func (t Manager) listenStd() {
	//tick := time.Tick(100 * time.Millisecond)

	for {
		select {
		case mem := <-t.MemPipe:
			fmt.Println(mem)
		case out := <-t.OutPipe:
			fmt.Println(out)
		case err := <-t.ErrPipe:
			fmt.Println(err)
		}
	}
}
