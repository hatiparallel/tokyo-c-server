package main

import (
	"time"
)

type Channel struct {
	Id   int64
	Name string
}

type Message struct {
	Id       int64
	Channel  int64
	Author   string
	IsEvent  int
	PostedAt time.Time
	Content  string
}
