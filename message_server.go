package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type MessageServer struct {
	listeners *Hub
}

func NewMessageServer() *MessageServer {
	return &MessageServer{NewHub()}
}

func (server *MessageServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	last_slash_index := strings.LastIndex(request.URL.Path, "/")

	if last_slash_index == -1 {
		writer.WriteHeader(404)
		fmt.Fprintln(writer, "not found")
		return
	}

	channel := request.URL.Path[last_slash_index+1:]

	if channel == "" {
		writer.WriteHeader(401)
		fmt.Fprintln(writer, "channel name must be designated")
		return
	}

	switch request.Method {
	case "GET":
		var (
			flushable bool
			flusher   http.Flusher
		)

		writer.Header().Set("Transfer-Encoding", "chunked")

		if flusher, flushable = writer.(http.Flusher); !flushable {
			return
		}

		listener := server.listeners.Subscribe(channel)

		defer server.listeners.Unsubscribe(listener)

		for {
			if payload, err := json.Marshal(<-listener); err == nil {
				fmt.Fprintf(writer, "%s\n", payload)
			}

			flusher.Flush()
		}
	case "POST":
		if request.Header.Get("Content-Type") != "application/json" {
			writer.WriteHeader(400)
			fmt.Fprintln(writer, "bad content type")
			return
		}

		var (
			err     error
			buffer  []byte
			message Message
		)

		if buffer, err = ioutil.ReadAll(request.Body); err != nil {
			writer.WriteHeader(400)
			fmt.Fprintln(writer, "invalid content stream")
			return
		}

		request.Body.Close()

		if json.Unmarshal(buffer, &message) != nil {
			writer.WriteHeader(400)
			fmt.Fprintln(writer, "corrupt content format")
			return
		}

		server.listeners.Publish(channel, message)

		writer.WriteHeader(201)
		fmt.Fprintf(writer, "Posted.\n")
	}
}
