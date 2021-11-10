package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func makeConnection(u url.URL, interrupt chan os.Signal) {
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			return
		}
	}

}

func createConnection(details *ConnDetails) {
	// Catch ctrl+c
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: details.Address, Path: "/ws"}
	if details.Duration > 0 {
		u.RawQuery = fmt.Sprintf("timeout=%d", details.Duration)
	}

	// Reconnect as many times as connections are set
	for i := 1; i <= details.Reconnects; i++ {
		makeConnection(u, interrupt)
	}
}

func launchParallelConnections(wg *sync.WaitGroup, details *ConnDetails) {
	for i := 1; i <= details.Parallel; i++ {
		wg.Add(1)
		go func() {
			createConnection(details)
			wg.Done()
		}()
	}
}

type ConnDetails struct {
	Address    string
	Duration   int
	Reconnects int
	Parallel   int
}

// Use cases
// - Serial connections in a row  (flag to control how many)
// - Multiple short-lived at once (flag to control how many)
// - Long term connection
// - Client controls duration via url param
func main() {
	addr := flag.String("addr", "localhost:8080", "http service address")
	duration := flag.Int("duration", 0, "duration in seconds of each connection. 0 is forver")
	reconnects := flag.Int("reconnects", 1, "how many times to reconnect befor quitting")
	parallel := flag.Int("parallel", 1, "how many connections to make in parallel")
	flag.Parse()

	details := &ConnDetails{
		Address:    *addr,
		Duration:   *duration,
		Reconnects: *reconnects,
		Parallel:   *parallel,
	}

	log.SetFlags(0)
	var wg sync.WaitGroup

	if details.Parallel > 1 {
		launchParallelConnections(&wg, details)
	}

	if details.Parallel == 1 {
		createConnection(details)
	}

	wg.Wait()
}
