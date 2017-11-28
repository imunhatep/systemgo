package system

import (
	"os/exec"
	"time"
	"io"
	"log"
	"fmt"
	"errors"
)

const UNIT_START_TIMEOUT = 10

type process struct {
	name    string
	cmd     *exec.Cmd
	Created time.Time
	Stopped time.Time
	Out     io.ReadCloser
	Err     io.ReadCloser
}

func NewProcess(name, target string, params []string) *process {
	process := process{}

	process.name = name
	process.cmd = exec.Command(target, params...)
	process.linkStd()

	return &process
}

func (t *process) Start(started chan<- error) {
	log.Printf("[%s] Starting...", t.name)

	if err := t.cmd.Start(); err != nil {
		log.Fatal(err)
	}

	var count int
	for t.cmd.Process == nil {
		time.Sleep(time.Second * 1)
		count += 1

		if count > UNIT_START_TIMEOUT {
			panic(fmt.Sprintf("[%s] PROCESS: Timedout(%d) waiting for process to start", t.name, UNIT_START_TIMEOUT))
		}
	}

	t.Created = time.Now()
	close(started)

	log.Printf("[%s] PROCESS PID: %d", t.name, t.GetPid())

	t.wait()
}

func (t process) Stop(done chan<- error) {
	if t.Finished() {
		log.Printf("[%s] PROCESS: Not Running", t.name)
	} else {
		// 5 seconds given to end process, or kill it
		select {
		case <-time.After(5 * time.Second):
			done <- t.kill()
		}
	}

	close(done)
}

func (t process) Running() bool {
	return t.cmd != nil && t.cmd.Process != nil && t.cmd.ProcessState == nil
}

func (t process) Finished() bool {
	return t.cmd == nil || t.cmd.ProcessState != nil
}

func (t process) GetName() string {
	return t.name
}

func (t process) GetPid() int {
	return t.GetCmd().Process.Pid
}

func (t process) GetCmd() *exec.Cmd {
	if !t.Running() {
		panic("Error in getting Command, exec is empty")
	}

	return t.cmd
}

func (t process) wait() {
	var stopped = make(chan error)

	go func() {
		stopped <- t.cmd.Wait()
	}()

	if err := <-stopped; err != nil {
		log.Printf("[%s] PROCESS: Finished with message: %s", t.name, err)
	} else {
		log.Printf("[%s] PROCESS: Finished", t.name)
	}
}

func (t process) kill() error {
	if t.Finished() {
		log.Printf("[%s] PROCESS: Not Running", t.name)
		return nil
	}

	var killErr string

	//if err := t.cmd.process.Signal(syscall.SIGINT); err != nil {
	if err := t.cmd.Process.Kill(); err != nil {
		killErr = fmt.Sprintf("[%s] PROCESS: Failed to kill PID [%d]: %s", t.name, t.GetPid(), err)
	} else {
		killErr = fmt.Sprintf("[%s] PROCESS: Killed by timeout", t.name)
	}

	return errors.New(killErr)
}

func (t *process) linkStd() {
	var err error

	t.Out, err = t.cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	t.Err, err = t.cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
}
