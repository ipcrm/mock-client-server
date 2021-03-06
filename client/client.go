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
	"github.com/ipcrm/mock-client-server/util"
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
		u.RawQuery = fmt.Sprintf("timeout=%f", details.Duration)
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
	Duration   float64
	Reconnects int
	Parallel   int
}

// Use cases
// - Serial connections in a row  (flag to control how many)
// - Multiple short-lived at once (flag to control how many)
// - Long term connection
// - Client controls duration via url param
func main() {
	addr := flag.String("addr",
		util.EnvString("ADDR", "localhost:8080"), util.HelpString("http service address", "ADDR"))
	duration := flag.Float64("duration",
		util.EnvFloat64("DURATION", 0.0), util.HelpString("duration in seconds (float) of each connection. 0 is forver (and the default)", "DURATION"))
	reconnects := flag.Int("reconnects", util.EnvInt("RECONNECTS", 1),
		util.HelpString("how many times to reconnect befor quitting", "RECONNECTS"))
	parallel := flag.Int("parallel",
		util.EnvInt("PARALLEL", 1), util.HelpString("how many connections to make in parallel", "PARALLEL"))
	flag.Parse()

	details := &ConnDetails{
		Address:    *addr,
		Duration:   *duration,
		Reconnects: *reconnects,
		Parallel:   *parallel,
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	var wg sync.WaitGroup
	launchParallelConnections(&wg, details)
	wg.Wait()
}
