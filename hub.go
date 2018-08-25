package main

import (
	"sync"
)

type Hub struct {
	mutex     *sync.RWMutex
	listeners map[chan Message]int64
	stamper   func(int64, *Message) error
}

func NewHub(stamper func(int64, *Message) error) *Hub {
	return &Hub{
		new(sync.RWMutex),
		make(map[chan Message]int64),
		stamper,
	}
}

func (hub *Hub) Stamp(channel int64, message *Message) error {
	return hub.stamper(channel, message)
}

func (hub *Hub) Subscribe(channel int64, listener chan Message) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	hub.listeners[listener] = channel

	return
}

func (hub *Hub) Publish(channel int64, message *Message) (err error) {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()

	err = hub.Stamp(channel, message)

	if err != nil {
		return err
	}

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
