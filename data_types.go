package main

import (
	"time"
)

type Person struct {
	Id   int64
	Name string
}

type Message struct {
	Id       int64
	Channel  int64
	Author   int64
	IsEvent  int
	PostedAt time.Time
	Content  string
}
