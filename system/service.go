package system

import (
	"time"
	"log"
	"io"
	"fmt"
	"bufio"
)

type Service struct {
	Name      string
	Exec      string
	Params    []string
	Restart   bool
	running   *process
	history   []*process
	isStarted bool
	isStopped bool
}

func (s Service) IsNew() bool {
	return len(s.history) == 0 && s.running == nil
}

func (s Service) IsRunning() bool {
	return s.running != nil && s.running.Running()
}

func (s Service) IsFinished() bool {
	// no running process, but have history, or process have exited
	return ( len(s.history) > 0 && s.running == nil) || ( s.running != nil && s.running.Finished() )
}

func (s Service) GetUsedMemory() uint64 {
	if !s.IsRunning() {
		return 0
	}

	mem, e := memoryUsage(s.running.GetPid())
	if e != nil {
		log.Println(e)
	}

	return mem
}

func (s *Service) Run(exit chan bool, out, err chan<- string) {
	if s.isStarted {
		log.Printf("[S][%s] Service already started", s.Name)
		return
	}

	s.isStopped = false
	s.isStarted = true

	// do not start process if Service is exit
	for !s.isStopped {
		select {
		case <-exit:
			s.stopProcess()
			return
		case <-time.After(time.Second):
			s.handleProcess(out, err)
		}
	}
}

func (s *Service) Stop(done chan<- error) {
	log.Printf("[S][%s] STOP exiting.. stopped: %d", s.Name, s.isStopped)

	time.Sleep(3 * time.Second)

	done <- nil
}

func (s Service) createProcess(out, err chan<- string) *process {
	running := NewProcess(s.Name, s.Exec, s.Params)

	started := make(chan error)
	go running.Start(started)
	<-started

	return running
}

func (s *Service) handleProcess(out, err chan<- string) {
	if s.IsNew() {
		log.Printf("[S][%s] creating new process", s.Name)
		s.running = s.createProcess(out, err)

		// listen for STD
		s.scanProcessStd(s.Name, &s.running.Out, out)
		s.scanProcessStd(s.Name, &s.running.Err, err)

		return
	}

	if s.IsFinished() {
		if s.running != nil {
			s.history = append(s.history, s.running)
			s.running = nil
		}

		if s.Restart {
			log.Printf("[S][%s] restarting", s.Name)
			s.running = s.createProcess(out, err)
			return
		}

		// nothing to do?!
		time.Sleep(3 * time.Second)
	}

	if s.IsRunning() {
		if time.Now().Second() % 10 == 0 {
			mem := s.GetUsedMemory()
			log.Printf("[S][%s][%d] memory usage: %.2d kb", s.Name, s.running.GetPid(), mem / 1024)
		}
	}
}

func (s *Service) stopProcess() error {
	if s.isStopped {
		log.Printf("[S][%s] Service.Stop() already have been called", s.Name)
		return nil
	}

	log.Printf("[S][%s] Received EXIT signal", s.Name)
	s.isStopped = true

	if s.IsRunning() {
		return s.running.Stop()
	}

	return nil
}

func (s Service) scanProcessStd(name string, pipe *io.ReadCloser, out chan<- string) {
	outScanner := bufio.NewScanner(*pipe)

	go func() {
		for s.IsRunning() && outScanner.Scan() {
			logs := outScanner.Text()
			out <- fmt.Sprintf("[%s] %s", name, logs)
		}
	}()
}
