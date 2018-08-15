package main

import (
	"sync"
	"time"
)

const (
	MESSAGE_TYPE_TEXT = iota
)

type Message struct {
	Id       int
	Author   int
	Type     int
	PostedAt time.Time
	Content  string
}

type Hub struct {
	mutex     *sync.RWMutex
	listeners map[chan Message]int
	journal   chan Message
}

func NewHub(journal chan Message) *Hub {
	return &Hub{
		new(sync.RWMutex),
		make(map[chan Message]int),
		journal,
	}
}

func (hub *Hub) Subscribe(channel int, listener chan Message) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	hub.listeners[listener] = channel

	return
}

func (hub *Hub) Publish(channel int, message Message) {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()

	for listener, _channel := range hub.listeners {
		if channel == _channel {
			listener <- message
		}

		if hub.journal != nil {
			hub.journal <- message
		}
	}
}

func (hub *Hub) Unsubscribe(listener chan Message) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	delete(hub.listeners, listener)
}
