package system

import (
	"log"
	"reflect"
	"fmt"
)

type Manager struct {
	taskList *[]Task

	MemPipe chan string
	OutPipe chan string
	ErrPipe chan string

	isRunning bool
	isStopped bool
}

func (t *Manager) Run(tasks *[]Task) {
	if t.isRunning || t.isStopped {
		log.Fatal("PM already started")
	}

	t.isStopped = false
	t.isRunning = true

	t.taskList = tasks

	bufSize := len(*tasks) * 2
	t.MemPipe = make(chan string, bufSize)
	t.OutPipe = make(chan string, bufSize)
	t.ErrPipe = make(chan string, bufSize)

	log.Println("Starting PM with:")

	var taskPipes []chan error

	for _, task := range *t.taskList {
		fmt.Printf("Name: %s\nExec: %s\nParams: %s\nRestart: %d\n", task.Name, task.Exec, task.Params, task.Restart)

		done := make(chan error, 1)
		taskPipes = append(taskPipes, done)

		go task.Run(done, t.OutPipe, t.ErrPipe)
	}

	go t.listenStd()

	log.Println("Waiting tasks to finish")
	t.waitFor(taskPipes)
}

func (t *Manager) Stop() {
	log.Println("Sending STOP to ALL")

	if !t.isRunning {
		log.Fatal("ProcessManager not started")
	}

	t.isStopped = true

	var tasks []chan error
	for _, task := range *t.taskList {
		done := make(chan error, 1)
		tasks = append(tasks, done)

		task.Stop(done)
	}

	log.Println("Waiting tasks to exit")
	t.waitFor(tasks)

	t.isRunning = false
}

func (t Manager) waitFor(tasks []chan error){
	cases := make([]reflect.SelectCase, len(tasks))
	for i, ch := range tasks {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}

	remaining := len(cases)
	for remaining > 0 {
		chosen, value, ok := reflect.Select(cases)
		if !ok {
			log.Printf("Read from channel %#v and received %s\n", tasks[chosen], value.String())
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
			log.Println(mem)
		case out := <-t.OutPipe:
			log.Println(out)
		case err := <-t.ErrPipe:
			log.Println(err)
		}
	}
}
