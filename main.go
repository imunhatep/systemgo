package main

import (
	"encoding/json"
	"io/ioutil"
	"github.com/imunhatep/systemg/system"
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"runtime"
	"log"
)

func main() {
	runtime.GOMAXPROCS(2)

	taskMng := &system.Manager{}
	go taskMng.Run(readConfig("./systemg.json"))

	sigChan := make(chan bool)
	handleSig(sigChan)
	<-sigChan

	fmt.Println("exiting")

	taskMng.Stop()
}

func handleSig(sigChan chan<- bool) {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		sigChan <- true
	}()

	fmt.Println("awaiting signal")
}

func readConfig(path string) []system.Service {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	var tasks []system.Service
	if err = json.Unmarshal(dat, &tasks); err != nil {
		log.Fatal(err)
	}

	return tasks
}
