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
	"runtime"
	"sync"
	"syscall"
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
	handleSig(&wg, sigChan)
	<-sigChan

	cancel()
	wg.Wait()
}

func handleSig(wg *sync.WaitGroup, sigChan chan<- bool) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigc:
			fmt.Println()
			fmt.Println(sig)
		case <-done:
			fmt.Println("All tasks are finished. Exiting..")
		}

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
