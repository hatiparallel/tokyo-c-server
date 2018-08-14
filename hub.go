package main

import (
	"sync"
	"time"
)

const (
	MESSAGE_TYPE_TEXT = iota
)

type Message struct {
	Id uint
	Author uint
	Type int
	PostedAt time.Time
	Content string
}

type Hub struct {
	mutex     *sync.RWMutex
	listeners map[chan Message]string
}

func NewHub() *Hub {
	return &Hub{new(sync.RWMutex), make(map[chan Message]string)}
}

func (hub *Hub) Subscribe(channel string) (listener chan Message) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	listener = make(chan Message)
	hub.listeners[listener] = channel

	return
}

func (hub *Hub) Publish(channel string, message Message) {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()

	for listener, _channel := range hub.listeners {
		if channel == _channel {
			listener <- message
		}
	}
}

func (hub *Hub) Unsubscribe(listener chan Message) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	delete(hub.listeners, listener)
}
