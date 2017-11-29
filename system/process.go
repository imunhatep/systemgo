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
	process := new(process)

	process.name = name
	process.cmd = exec.Command(target, params...)

	var err error
	process.Out, err = process.cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	process.Err, err = process.cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	return process
}

func (t *process) Start(started chan<- error) {
	log.Printf("[PROCESS][%s] Starting...", t.name)

	if err := t.cmd.Start(); err != nil {
		log.Fatal(err)
	}

	var count int
	for t.cmd.Process == nil {
		time.Sleep(time.Second * 1)
		count += 1

		if count > UNIT_START_TIMEOUT {
			panic(fmt.Sprintf("[PROCESS][%s] Timedout(%d) waiting for process to start", t.name, UNIT_START_TIMEOUT))
		}
	}

	t.Created = time.Now()
	close(started)

	log.Printf("[PROCESS][%s] PID: %d", t.name, t.GetPid())

	t.wait()
}

func (t process) Stop() error {
	if t.Finished() && !t.Running() {
		log.Printf("[PROCESS][%s] Not Running", t.name)
	} else {
		log.Printf("[PROCESS][%s] Killing..", t.name)

		// 3 seconds given to end process, or kill it
		select {
		case <-time.After(3 * time.Second):
			return t.kill()
		}
	}

	return nil
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

func (t *process) wait() {
	var stopped = make(chan error)

	go func() { stopped <- t.cmd.Wait() }()

	if err := <-stopped; err != nil {
		log.Printf("[PROCESS][%s] Finished with message: %s", t.name, err)
	} else {
		log.Printf("[PROCESS][%s] Finished", t.name)
	}

	t.Stopped = time.Now()
}

func (t *process) kill() error {
	if t.Finished() {
		log.Printf("[PROCESS][%s] Nothing to kill", t.name)
		return nil
	}

	var killErr string

	//if err := t.cmd.Process.Signal(syscall.SIGINT); err != nil {
	if err := t.cmd.Process.Kill(); err != nil {
		killErr = fmt.Sprintf("[PROCESS][%s] Failed to kill PID [%d]: %s", t.name, t.GetPid(), err)
	} else {
		killErr = fmt.Sprintf("[PROCESS][%s] Killed by timeout", t.name)
		t.Stopped = time.Now()
	}

	return errors.New(killErr)
}
