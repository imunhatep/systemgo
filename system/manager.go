package system

import (
	"log"
	"reflect"
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
		log.Fatal("TaskManager already started")
	}

	log.Printf("Starting PM with: %s", tasks)

	t.isStopped = false
	t.isRunning = true

	t.taskList = tasks

	bufSize := len(*tasks) * 2
	t.MemPipe = make(chan string, bufSize)
	t.OutPipe = make(chan string, bufSize)
	t.ErrPipe = make(chan string, bufSize)

	for _, task := range *t.taskList {
		started := make(chan error)
		go task.Run(started, t.OutPipe, t.ErrPipe)
		<-started
	}
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

	log.Println("Validating STOPPED tasks")
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

	t.isRunning = false
}