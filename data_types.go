package main

import (
	"sync"
	"time"
)

type http_status struct {
	code int
	body interface{}
}

type pin_ticket struct {
	pin      int
	owner    string
	pendings map[string]bool
	channel  chan string
	mutex    *sync.Mutex
}

type PINEvent struct {
	Type   string
	PIN    int
	Person string
}

type Channel struct {
	Id      int
	Name    string
	Members []string
}

type Message struct {
	Id       int
	Channel  int
	Author   string
	IsEvent  int
	PostedAt time.Time
	Content  string
}

type Summary struct {
	ChannelId   int
	ChannelName string
	MessageId   int
}

type Status struct {
	FriendshipCount int
	Latests         []Summary
}
