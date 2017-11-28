package main

import (
	"encoding/json"
	"io/ioutil"
	"github.com/imunhatep/systemg/system"
	"os"
	"os/signal"
	"syscall"
	"fmt"
)

func main() {
	taskMng := system.Manager{}
	go taskMng.Run(readConfig("./systemg.json"))

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

func readConfig(path string) *[]system.Task {
	dat, err := ioutil.ReadFile(path)
	check(err)

	var tasks []system.Task
	err = json.Unmarshal(dat, &tasks)

	check(err)

	return &tasks
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
