package main

import (
	"encoding/json"
	"flag"
	"github.com/imunhatep/systemgo/system"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(*flag.Int("j", 2, "GOMAXPROCS"))

	taskMng := &system.Manager{}
	go taskMng.Run(readConfig(*flag.String("f", "tasks.json", "JSON file with defined tasks")))

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
