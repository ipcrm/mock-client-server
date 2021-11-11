package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Reader for long-term conns
func reader(conn *websocket.Conn, id uuid.UUID) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("%s: Client %s still connected", time.Now().Format("2006-01-02 15:04:05.000"), id)

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}

	}
}

// Reader for short-term conns
func readerShortLived(conn *websocket.Conn, timeout int, id uuid.UUID) {
	timeOut := time.Duration(timeout) * time.Second
	queryCtx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	stayOpen := true
	for stayOpen {
		select {
		case <-queryCtx.Done():
			log.Printf("Timeout expired. Disconnecting client %s\n", id)
			err := conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "time's up; goodbye!"),
				time.Now().Add(time.Second),
			)
			stayOpen = false
			if err != nil {
				log.Println(err)
				return
			}
		default:
			// read in a message
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			log.Printf("Client %s still connected\n", id)
			if err := conn.WriteMessage(messageType, p); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	timeout := r.URL.Query()["timeout"]
	id := uuid.New()

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Printf("Client %s Connected\n", id)
	err = ws.WriteMessage(1, []byte(fmt.Sprintf("Connected! Client ID: %s", id)))
	if err != nil {
		log.Println(err)
	}
	if timeout != nil {
		timeoutInt, err := strconv.Atoi(timeout[0])
		if err != nil {
			w.WriteHeader(500)
			if _, err := w.Write([]byte("Cannot parse timeout value!")); err != nil {
				log.Println(err)
				return
			}
			return
		}
		err = ws.WriteMessage(1, []byte(fmt.Sprintf("Connection will be closed in %s seconds!", timeout[0])))
		if err != nil {
			log.Println(err)
		}
		readerShortLived(ws, timeoutInt, id)
	}

	if timeout == nil {
		reader(ws, id)
	}
}

func setupRoutes() {
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	port := flag.Int("port", 8080, "http service port")
	addr := flag.String("addr", "*", "http service address")
	flag.Parse()

	serviceAddress := fmt.Sprintf("%s:%d", *addr, *port)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("Starting server on %s\n", serviceAddress)
	setupRoutes()
	log.Fatal(http.ListenAndServe(strings.ReplaceAll(serviceAddress, "*", ""), nil))
}
