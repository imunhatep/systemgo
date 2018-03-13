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

func (p *process) Start(started chan<- error) {
	log.Printf("[P][%s] Starting...", p.name)

	if err := p.cmd.Start(); err != nil {
		log.Fatal(err)
	}

	var count int
	for p.cmd.Process == nil {
		time.Sleep(time.Second * 1)
		count += 1

		if count > UNIT_START_TIMEOUT {
			panic(fmt.Sprintf("[P][%s] Timedout(%d) waiting for process to start", p.name, UNIT_START_TIMEOUT))
		}
	}

	p.Created = time.Now()
	close(started)

	log.Printf("[P][%s] PID: %d", p.name, p.GetPid())

	p.wait()
}

func (p process) Stop() error {
	if p.Finished() && !p.Running() {
		log.Printf("[P][%s] Not running", p.name)
	} else {
		log.Printf("[P][%s] Killing..", p.name)

		// 10 seconds given to end process, or kill it
		select {
		case <-time.After(UNIT_START_TIMEOUT * time.Second):
			return p.kill()
		}
	}

	return nil
}

func (p process) Running() bool {
	return p.cmd != nil && p.cmd.Process != nil && p.cmd.ProcessState == nil
}

func (p process) Finished() bool {
	return p.cmd == nil || p.cmd.ProcessState != nil
}

func (p process) GetName() string {
	return p.name
}

func (p process) GetPid() int {
	return p.GetCmd().Process.Pid
}

func (p process) GetCmd() *exec.Cmd {
	if !p.Running() {
		panic("Error in getting Command, exec is empty")
	}

	return p.cmd
}

func (p *process) wait() {
	var stopped = make(chan error)

	go func() { stopped <- p.cmd.Wait() }()

	if err := <-stopped; err != nil {
		log.Printf("[P][%s] Finished with message: %s", p.name, err)
	} else {
		log.Printf("[P][%s] Finished", p.name)
	}

	p.Stopped = time.Now()
}

func (p *process) kill() error {
	if p.Finished() {
		log.Printf("[P][%s] Nothing to kill", p.name)
		return nil
	}

	var killErr string

	//if err := m.cmd.Process.Signal(syscall.SIGINT); err != nil {
	if err := p.cmd.Process.Kill(); err != nil {
		killErr = fmt.Sprintf("[P][%s] Failed to kill PID [%d]: %s", p.name, p.GetPid(), err)
	} else {
		killErr = fmt.Sprintf("[P][%s] Killed by timeout", p.name)
		p.Stopped = time.Now()
	}

	return errors.New(killErr)
}
