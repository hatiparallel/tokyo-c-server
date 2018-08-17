package main

import (
	"sync"
)

type Channel int
type Message map[string]interface{}

type Hub struct {
	mutex     *sync.RWMutex
	listeners map[chan Message]Channel
	stamper   func(Channel, Message) error
}

func NewHub(stamper func(Channel, Message) error) *Hub {
	return &Hub{
		new(sync.RWMutex),
		make(map[chan Message]Channel),
		stamper,
	}
}

func (hub *Hub) Stamp(channel Channel, message Message) error {
	return hub.stamper(channel, message)
}

func (hub *Hub) Subscribe(channel Channel, listener chan Message) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	hub.listeners[listener] = channel

	return
}

func (hub *Hub) Publish(channel Channel, message Message) (err error) {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()

	err = hub.Stamp(channel, message)

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
