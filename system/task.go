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
	// no running process, but have history, or process have exited
	return ( len(t.history) > 0 && t.running == nil) || ( t.running != nil && t.running.Finished() )
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

func (t *Task) Run(finished chan<- error, out, err chan<- string) {
	if t.isStarted {
		log.Printf("[%s] Task already started", t.Name)
		return
	}

	t.isStopped = false
	t.isStarted = true

	for !t.isStopped {
		log.Printf("\n%s \nnew: %d, fin: %d, run: %d\n\n", t.Name, t.IsNew(), t.IsFinished(), t.IsRunning())

		// do not start process if Task is stopped
		if t.IsNew() {
			log.Printf("[%s] process creating..", t.Name)
			t.running = t.execProcess(out, err)
		} else if t.IsFinished() {
			if t.running != nil {
				t.history = append(t.history, t.running)
				t.running = nil
			}

			// do not re-start process if Task is stopped
			if t.Restart {
				log.Printf("[%s] process restarting..", t.Name)
				t.running = t.execProcess(out, err)
			}

			log.Printf("[%s] history count: %d", t.Name, len(t.history))
		} else if t.IsRunning() {
			log.Printf("[%s][%d] process is running", t.running.GetName(), t.running.GetPid())
		} else {
			log.Printf("[%s] nothing to do, stopped: %s", t.running.GetName(), t.isStopped)
		}

		time.Sleep(time.Second)
	}

	close(finished)
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

func (t Task) execProcess(out, err chan<- string) *process {
	running := NewProcess(t.Name, t.Exec, t.Params)

	started := make(chan error)
	go running.Start(started)
	<-started

	t.scanProcessStd(t.Name, running.Out, out)
	t.scanProcessStd(t.Name, running.Err, err)

	return running
}

func (t Task) scanProcessStd(name string, pipe io.ReadCloser, out chan<- string) {
	outScanner := bufio.NewScanner(pipe)

	go func() {
		for outScanner.Scan() {
			logs := outScanner.Text()
			out <- fmt.Sprintf("[%s][STD] %s", name, logs)
		}
	}()
}
