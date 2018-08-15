package main

import (
	"sync"
)

type Message map[string]interface{}

type Hub struct {
	mutex     *sync.RWMutex
	listeners map[chan Message]int
	stamper   func(Message) error
}

func NewHub(stamper func(Message) error) *Hub {
	return &Hub{
		new(sync.RWMutex),
		make(map[chan Message]int),
		stamper,
	}
}

func (hub *Hub) Stamp(message Message) error {
	return hub.stamper(message)
}

func (hub *Hub) Subscribe(channel int, listener chan Message) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	hub.listeners[listener] = channel

	return
}

func (hub *Hub) Publish(channel int, message Message) (err error) {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()

	err = hub.Stamp(message)

	if err != nil {
		return err
	}

	for listener, _channel := range hub.listeners {
		if channel == _channel {
			listener <- message
		}
	}

	return
}

func (hub *Hub) Unsubscribe(listener chan Message) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	delete(hub.listeners, listener)
}
