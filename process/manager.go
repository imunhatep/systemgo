package process

import (
	"bufio"
	"log"
	"fmt"
	"reflect"
	"io"
	"time"
)

type Manager struct {
	taskList *[]Task
	running  map[string]*Task
	stopped  map[string]*Task

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
	t.isStopped = false
	t.isRunning = true

	t.taskList = tasks
	t.running = make(map[string]*Task)
	t.stopped = make(map[string]*Task)

	t.MemPipe = make(chan string, 10)
	t.OutPipe = make(chan string, 100)
	t.ErrPipe = make(chan string, 100)

	go func() {
		for {
			if t.isStopped {
				return
			}

			t.monitor()
			time.Sleep(time.Second)
		}
	}()
}

func (t *Manager) monitor() {
	for _, task := range *t.taskList {
		taskProcess, ok := t.running[task.Name]

		if !ok {
			log.Printf("[%s] creating process", task.Name)
			taskProcess = &Task{Name: task.Name, Exec: task.Exec, Params: task.Params}
		} else {
			log.Printf("[%s] process exists", taskProcess.Name)
		}

		if taskProcess.Finished() {
			_, beenRunning := t.stopped[task.Name]

			if taskProcess.Restart || !beenRunning {
				log.Printf("[%s] starting process", taskProcess.Name)
				t.running[taskProcess.Name] = taskProcess
				t.exec(taskProcess)
			} else {
				log.Printf("[%s] removing process", taskProcess.Name)

				t.stopped[taskProcess.Name] = taskProcess
				delete(t.running, taskProcess.Name)
			}
		}
	}
}

func (t Manager) exec(taskProcess *Task) {
	go taskProcess.Start()
	t.checkMemory(taskProcess, t.MemPipe)
	t.scanStd(taskProcess.Name, taskProcess.out, t.OutPipe)
	t.scanStd(taskProcess.Name, taskProcess.err, t.ErrPipe)
}

func (t *Manager) Stop() {
	log.Println("Sending STOP to ALL")

	if !t.isRunning {
		log.Fatal("ProcessManager not started")
	}

	t.isStopped = true

	var tasks []chan error
	for _, task := range t.running {
		done := make(chan error, 1)
		tasks = append(tasks, done)

		task.Stop(done)
	}

	log.Println("Validating STOPPED")
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

func (t Manager) checkMemory(taskProcess *Task, memout chan<- string) {
	log.Print(fmt.Sprintf("[%s] Memory...\t\t\t", taskProcess.Name))

	tick := time.Tick(2 * time.Second)

	go func() {
		for {
			select {
			case <-tick:

				mem, e := memoryUsage(taskProcess.GetPid())

				if e != nil {
					if taskProcess.Finished() {
						return
					} else {
						log.Println(e)
					}
				}

				// check this type assertion to avoid a panic
				memout <- fmt.Sprintf("[%s] MEM[%d]: %d bytes", taskProcess.Name, taskProcess.GetPid(), mem)
			}
		}
	}()
}

func (t Manager) scanStd(name string, pipe io.ReadCloser, out chan<- string) {
	outScanner := bufio.NewScanner(pipe)

	go func() {
		for outScanner.Scan() {
			logs := outScanner.Text()
			out <- fmt.Sprintf("[%s] %s", name, logs)
		}
	}()
}
