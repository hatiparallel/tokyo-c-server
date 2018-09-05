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
	subject, err := authenticate(request)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	switch request.Method {
	case "GET":
		pin_table.mutex.Lock()

		ticket, exists := pin_table.by_owner[subject]

		if !exists {
			ticket = issue_ticket(subject)
			pin_table.by_pin[ticket.pin] = ticket
			pin_table.by_owner[subject] = ticket
		}

		pin_table.mutex.Unlock()

		defer func() {
			close(ticket.channel)

			pin_table.mutex.Lock()
			delete(pin_table.by_owner, subject)
			delete(pin_table.by_pin, ticket.pin)
			pin_table.mutex.Unlock()
		}()

		writer.Header().Set("Transfer-Encoding", "chunked")

		flusher, flushable := writer.(http.Flusher)

		if !flushable {
			return &http_status{400, "streaming cannot be established"}
		}

		encoder := json.NewEncoder(writer)
		ticker := time.Tick(3 * time.Second)

		encoder.Encode(&pin_event{"pin", ticket.pin, ""})
		flusher.Flush()

	LOOP:
		for {
			select {
			case <-ticker:
				if encoder.Encode(&pin_event{"noop", 0, ""}) != nil {
					break LOOP
				}
			case person_id := <-ticket.channel:
				if encoder.Encode(&pin_event{"request", 0, person_id}) != nil {
					break LOOP
				}
			}

			flusher.Flush()
		}

		return nil
	default:
		return &http_status{405, "method not allowed"}
	}
}

func issue_ticket(owner string) *pin_ticket {
	var pin int

	rand.Seed(time.Now().UnixNano())

	for {
		pin = int(10000000 + rand.Int31()%90000000)

		if _, exists := pin_table.by_pin[pin]; !exists {
			break
		}
	}

	ticket := new(pin_ticket)
	ticket.pin = pin
	ticket.owner = owner
	ticket.pendings = make(map[string]bool)
	ticket.channel = make(chan string)
	ticket.mutex = new(sync.Mutex)

	return ticket
}
