package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/imunhatep/systemgo/system"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"runtime"
	"sync"
)

func main() {
	runtime.GOMAXPROCS(*flag.Int("j", 2, "GOMAXPROCS"))

	taskList := readConfig(*flag.String("f", "tasks.json", "JSON file with defined tasks"))
	serviceMng := system.NewServiceManager(taskList)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		serviceMng.Run(ctx)

		wg.Done()
	}()

	sigChan := make(chan bool)
	handleSig(sigChan)
	<-sigChan

	cancel()
	wg.Wait()
}

func handleSig(sigChan chan<- bool) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigc
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
