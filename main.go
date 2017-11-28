package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"github.com/imunhatep/systemg/process"
	"os"
	"os/signal"
	"syscall"
	"fmt"
)

func main() {
	taskMng := process.Manager{}
	taskMng.Run(readConfig("./systemg.json"))

	go showStatus(taskMng.MemPipe, taskMng.OutPipe, taskMng.ErrPipe)

	sigChan := make(chan bool)
	handleSig(sigChan)
	<-sigChan

	fmt.Println("exiting")


	taskMng.Stop()
}

func handleSig(sigChan chan<- bool) {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		sigChan <- true
	}()

	fmt.Println("awaiting signal")
}

func showStatus(memout, stdout, stderr <-chan string) {
	//tick := time.Tick(100 * time.Millisecond)

	for {
		select {
		case mem := <-memout:
			log.Println(mem)
		case out := <-stdout:
			log.Println(out)
		case err := <-stderr:
			log.Println(err)
		}
	}
}

func readConfig(path string) *[]process.Task {
	dat, err := ioutil.ReadFile(path)
	check(err)

	var tasks []process.Task
	err = json.Unmarshal(dat, &tasks)

	check(err)

	return &tasks
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
