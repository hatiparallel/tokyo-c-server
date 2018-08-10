package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type Message struct {
	User   int
	Text   string
	Posted int
}


type Hub struct {
	mutex *sync.RWMutex
	listeners []chan Message
}

func NewHub() *Hub {
	return &Hub{new(sync.RWMutex), make([]chan Message, 0, 16)}
}

func (hub *Hub) AddListener() (listener chan Message) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	listener = make(chan Message)
	hub.listeners = append(hub.listeners, listener)

	return
}

func (hub *Hub) Broadcast(message Message) {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()

	for _, listener := range hub.listeners {
		listener <- message
	}
}

func (hub *Hub) RemoveListener(target chan Message) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	listeners := make([]chan Message, 0, len(hub.listeners))

	for _, listener := range(hub.listeners) {
		if (listener != target) {
			listeners = append(listeners, listener)
		}
	}

	hub.listeners = listeners
}

func main() {
	listeners := NewHub()

	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			var (
				flushable bool
				flusher   http.Flusher
			)

			header := w.Header()
			header.Set("Transfer-Encoding", "chunked")

			if flusher, flushable = w.(http.Flusher); !flushable {
				return
			}

			listener := listeners.AddListener()

			defer listeners.RemoveListener(listener)

			for {
				if  payload, err := json.Marshal(<-listener); err == nil {
					fmt.Fprintf(w, "%s\n", payload)
				}

				flusher.Flush()
			}
		case "POST":
			if r.Header.Get("Content-Type") != "application/json" {
				w.WriteHeader(400)
				fmt.Fprintln(w, "bad content type")
				return
			}

			var (
				err       error
				buffer    []byte
				read_size int
				message   Message
			)

			buffer = make([]byte, 4096)

			if read_size, err = r.Body.Read(buffer); err != io.EOF {
				w.WriteHeader(400)
				fmt.Fprintln(w, "invalid length of content")
				return
			}

			r.Body.Close()

			if json.Unmarshal(buffer[0:read_size], &message) != nil {
				w.WriteHeader(400)
				fmt.Fprintln(w, "corrupt content format")
				return
			}

			listeners.Broadcast(message)

			fmt.Fprintf(w, "Done.\n")
		}
	})

	http.ListenAndServe(":80", nil)
}
