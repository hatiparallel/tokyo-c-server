package main

import (
	"encoding/json"
	"time"
	"math/rand"
	"net/http"
	"sync"
)

type pin_event struct {
	Type string
	PIN int
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

		if pin, exists := pin_table.inverse[subject]; exists {
			delete(pin_table.inverse, subject)
			delete(pin_table.store, pin)
		}

		rand.Seed(time.Now().UnixNano())

		pin_table.mutex.Lock()

		var pin int

		for {
			pin = int(10000000 + rand.Int31() % 90000000)

			if _, exists := pin_table.store[pin]; !exists {
				break
			}
		}

		ticket := pin_table.store[pin]
		ticket.mutex = new(sync.Mutex)
		ticket.mutex.Lock()

		pin_table.mutex.Unlock()

		defer delete(pin_table.store, pin)
		
		ticket.owner = subject
		ticket.channel = make(chan string)
		ticket.mutex.Unlock()

		encoder := json.NewEncoder(writer)

		encoder.Encode(&pin_event{"pin", pin, ""})

		for {
			if encoder.Encode(&pin_event{"request", 0, <-ticket.channel}) != nil {
				break
			}
		}

		return nil
	default:
		return &http_status{405, "method not allowed"}
	}
}
