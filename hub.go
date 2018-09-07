package main

import (
	"sync"
)

type Hub struct {
	mutex     *sync.RWMutex
	listeners map[chan Message]int
}

func NewHub() *Hub {
	return &Hub{
		new(sync.RWMutex),
		make(map[chan Message]int),
	}
}

func (hub *Hub) Subscribe(channel int, listener chan Message) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	hub.listeners[listener] = channel

	return
}

func (hub *Hub) Publish(channel int, message *Message) {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()

	for listener, _channel := range hub.listeners {
		if channel == _channel {
			listener <- *message
		}
	}

	return
}

func (hub *Hub) Unsubscribe(listener chan Message) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	delete(hub.listeners, listener)
}
