package system

import (
	"time"
	"log"
	"io"
	"fmt"
	"bufio"
)

type Task struct {
	Name      string
	Exec      string
	Params    []string
	Restart   bool
	running   *process
	history   []*process
	isStarted bool
	isStopped bool
}

func (t Task) IsNew() bool {
	return len(t.history) == 0 && t.running == nil
}

func (t Task) IsRunning() bool {
	return t.running != nil && t.running.Running()
}

func (t Task) IsFinished() bool {
	return len(t.history) > 0 && ( t.running == nil || t.running.Finished() )
}

func (t Task) GetUsedMemory(memout chan<- string) {
	if !t.IsRunning() {
		return
	}

	mem, e := memoryUsage(t.running.GetPid())

	if e != nil {
		log.Println(e)
	}

	// check this type assertion to avoid a panic
	memout <- fmt.Sprintf("[%s] TASK PID[%d] memory usage: %d bytes", t.Name, t.running.GetPid(), mem)
}

func (t *Task) Run(started chan<- error, out, err chan<- string) {

	if t.isStarted {
		log.Println("Task is running")
		return
	}

	t.isStopped = false
	t.isStarted = true

	for {
		log.Println(t.Name, "new: ", t.IsNew(), "fin: ", t.IsFinished(), "run: ", t.IsRunning())

		// do not start process if Task is stopped
		if t.IsNew() && !t.isStopped {
			log.Printf("[%s] process creating..", t.Name)
			t.execProcess(started, out, err)
		} else if t.IsFinished() {
			t.history = append(t.history, t.running)
			t.running = nil

			// do not re-start process if Task is stopped
			if t.Restart && t.isStopped {
				log.Printf("[%s] process restarting..", t.Name)

				started := make(chan error)
				t.execProcess(started, out, err)
				<-started
			}
		} else if t.IsRunning() {
			log.Printf("[%s] process is running", t.running.GetName())
		} else {
			log.Printf("[%s] nothing to do, stopped: %s", t.running.GetName(), t.isStopped)
		}

		time.Sleep(time.Second)
	}
}

func (t *Task) Stop(done chan<- error) {
	log.Printf("[%s] STOP!", t.Name)

	if !t.isStarted {
		log.Fatal("Task is not running")
		return
	}

	if t.isStopped {
		log.Printf("[%s] Task.Stop() already have been called", t.Name)
		return
	}

	t.isStopped = true

	if t.running != nil {
		t.running.Stop(done)
	} else {
		close(done)
	}

	t.isStarted = false
}

func (t *Task) execProcess(done chan<- error, out, err chan<- string) {
	t.running = NewProcess(t)

	started := make(chan error)
	go t.running.Start(started)
	<-started

	t.scanProcessStd(t.Name, t.running.Out, out)
	t.scanProcessStd(t.Name, t.running.Err, err)

	close(done)
}

func (t Task) scanProcessStd(name string, pipe io.ReadCloser, out chan<- string) {
	outScanner := bufio.NewScanner(pipe)

	go func() {
		for outScanner.Scan() {
			logs := outScanner.Text()
			out <- fmt.Sprintf("[%s] %s", name, logs)
		}
	}()
}
