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
	memout <- fmt.Sprintf("[TASK][%s] PID[%d] memory usage: %d bytes", t.Name, t.running.GetPid(), mem)
}

func (t *Task) Run(exit chan bool, out, err chan<- string) {
	if t.isStarted {
		log.Printf("[TASK][%s] Task already started", t.Name)
		return
	}

	t.isStopped = false
	t.isStarted = true

	// do not start process if Task is exit
	for !t.isStopped {
		select {
		case <-exit:
			t.isStopped = true
			t.stopProcess()
			return
		case <-time.After(time.Second):
			t.handleProcess(out, err)
		}
	}
}

func (t *Task) Stop(done chan<- error) {
	log.Printf("[TASK][%s] STOP exiting.. stopped: %d", t.Name, t.isStopped)

	time.Sleep(3 * time.Second)

	done <- nil
}

func (t Task) createProcess(out, err chan<- string) *process {
	running := NewProcess(t.Name, t.Exec, t.Params)

	t.scanProcessStd(t.Name, &running.Out, out)
	t.scanProcessStd(t.Name, &running.Err, err)

	started := make(chan error)
	go running.Start(started)
	<-started

	return running
}

func (t *Task) handleProcess(out, err chan<- string) {
	if t.IsNew() {
		log.Printf("[TASK][%s] creating", t.Name)
		t.running = t.createProcess(out, err)

		return
	}

	if t.IsFinished() {
		if t.running != nil {
			t.history = append(t.history, t.running)
			t.running = nil
		}

		if t.Restart {
			log.Printf("[TASK][%s] restarting", t.Name)
			t.running = t.createProcess(out, err)
			return
		}

		// nothing to do?!
		time.Sleep(4 * time.Second)
	}

	if t.IsRunning() {
		//log.Printf("[TASK][%s][%d] running with stopped?: %d", t.running.GetName(), t.running.GetPid(), t.isStopped)
		return
	}
}

func (t *Task) stopProcess() error {
	if t.isStopped {
		log.Printf("[TASK][%s] Task.Stop() already have been called", t.Name)
		return nil
	}

	log.Printf("[TASK][%s] Received EXIT signal", t.Name)
	t.isStopped = true

	if t.IsRunning() {
		return t.running.Stop()
	}

	return nil
}

func (t Task) scanProcessStd(name string, pipe *io.ReadCloser, out chan<- string) {
	outScanner := bufio.NewScanner(*pipe)

	go func() {
		for outScanner.Scan() {
			logs := outScanner.Text()
			out <- fmt.Sprintf("[%s] %s", name, logs)
		}
	}()
}
