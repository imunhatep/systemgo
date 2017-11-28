package process

import (
	"time"
	"log"
	"os/exec"
	"io"
	"fmt"
	"errors"
)

type History struct{}

type Task struct {
	Name      string
	Exec      string
	Params    []string
	Restart   bool
	isRunning bool
	executes  []*exec.Cmd
	created   time.Time
	out       io.ReadCloser
	err       io.ReadCloser
}

func (t Task) GetPid() int {
	if len(t.executes) < 1 {
		return -1
	}

	return t.getCmd().Process.Pid
}

func (t Task) getCmd() *exec.Cmd {
	if len(t.executes) < 1 {
		panic("Error in getting last executable Command, exec history is empty")
	}

	return t.executes[len(t.executes)-1]
}

func (t Task) Running() bool {
	return t.isRunning
}

func (t *Task) Start() {
	log.Printf("[%s] Starting...", t.Name)

	cmd := exec.Command(t.Exec, t.Params...)
	t.executes = append(t.executes, cmd)
	t.linkStdOutputs()

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	var count int
	for cmd.Process == nil {
		time.Sleep(time.Second * 1)
		count += 1

		if count > 10 {
			log.Println("Timedout waiting for process to start, trying to start anyway")
		}
	}

	t.isRunning = true
	t.created = time.Now()

	log.Printf("[%s] Started at PID: %d", t.Name, t.GetPid())

	t.onExit()
}

func (t Task) Stop(done chan<- error) {
	if t.Finished() {
		log.Printf("[%s] Not Running", t.Name)
		done <- nil
		close(done)
		return
	}

	// 5 seconds given to end process, or kill it
	select {
	case <-time.After(5 * time.Second):
		done <- t.Kill()
	}

	close(done)
}

func (t *Task) Kill() error {
	if t.Finished() {
		log.Printf("[%s] Not Running", t.Name)
		return nil
	}

	var killErr string

	//if err := t.cmd.Process.Signal(syscall.SIGINT); err != nil {
	if err := t.getCmd().Process.Kill(); err != nil {
		killErr = fmt.Sprintf("[%s] Failed to kill PID [%d]: %s", t.Name, t.GetPid(), err)
	} else {
		t.isRunning = false
		killErr = fmt.Sprintf("[%s] Process killed by timeout", t.Name)
	}

	return errors.New(killErr)
}

func (t Task) Finished() bool {
	if len(t.executes) < 1 {
		return true
	}

	cmd := t.getCmd()
	return cmd == nil || cmd.ProcessState != nil
}

func (t *Task) linkStdOutputs() {
	var err error

	t.out, err = t.getCmd().StdoutPipe()
	if err != nil {
		panic(err)
	}

	t.err, err = t.getCmd().StderrPipe()
	if err != nil {
		panic(err)
	}
}

func (t *Task) onExit() {
	var stopped = make(chan error)

	go func() {
		stopped <- t.getCmd().Wait()
	}()

	if err := <-stopped; err != nil {
		log.Printf("[%s] Finished with message: %s", t.Name, err)
	} else {
		log.Printf("[%s] Exited", t.Name)
	}

	t.isRunning = false
}
