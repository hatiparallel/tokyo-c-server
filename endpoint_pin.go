package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type pin_ticket struct {
	pin      int
	owner    string
	pendings map[string]bool
	channel  chan string
	mutex    *sync.Mutex
}

type pin_event struct {
	Type   string
	PIN    int
	Person string
}

func endpoint_pin(writer http.ResponseWriter, request *http.Request) *http_status {
	var subject string

	err := authenticate(request, &subject)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	switch request.Method {
	case "GET":
		pin_table.mutex.Lock()

		if ticket, exists := pin_table.by_owner[subject]; exists {
			delete(pin_table.by_owner, subject)
			delete(pin_table.by_pin, ticket.pin)
		}

		rand.Seed(time.Now().UnixNano())

		var pin int

		for {
			pin = int(10000000 + rand.Int31()%90000000)

			if _, exists := pin_table.by_pin[pin]; !exists {
				break
			}
		}

		ticket := new(pin_ticket)
		ticket.pin = pin
		ticket.owner = subject
		ticket.pendings = make(map[string]bool)
		ticket.channel = make(chan string)
		ticket.mutex = new(sync.Mutex)
		pin_table.by_pin[pin] = ticket
		pin_table.by_owner[subject] = ticket
		pin_table.mutex.Unlock()

		writer.Header().Set("Transfer-Encoding", "chunked")

		flusher, flushable := writer.(http.Flusher)

		if !flushable {
			return &http_status{400, "streaming cannot be established"}
		}

		encoder := json.NewEncoder(writer)

		encoder.Encode(&pin_event{"pin", pin, ""})
		flusher.Flush()

		for {
			if encoder.Encode(&pin_event{"request", 0, <-ticket.channel}) != nil {
				break
			}

			flusher.Flush()
		}

		return nil
	default:
		return &http_status{405, "method not allowed"}
	}
}
