package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type MessageServer struct {
	listeners *Hub
}

func NewMessageServer(stamper func(int64, *Message) error) *MessageServer {
	return &MessageServer{NewHub(stamper)}
}

func (server *MessageServer) handle_request(writer http.ResponseWriter, request *http.Request) *http_status {
	channel, err := strconv.Atoi(strings.TrimPrefix(request.URL.Path, "/streams/"))

	if err != nil {
		return &http_status{400, "invalid channel id"}
	}

	switch request.Method {
	case "GET":
		var (
			flushable bool
			flusher   http.Flusher
		)

		writer.Header().Set("Transfer-Encoding", "chunked")

		if flusher, flushable = writer.(http.Flusher); !flushable {
			return &http_status{400, "streaming cannot be established"}
		}

		listener := make(chan Message)

		server.listeners.Subscribe(int64(channel), listener)

		defer server.listeners.Unsubscribe(listener)

		for {
			if payload, err := json.Marshal(<-listener); err == nil {
				fmt.Fprintf(writer, "%s\n", payload)
			}

			flusher.Flush()
		}
	case "POST":
		if request.Header.Get("Content-Type") != "application/json" {
			return &http_status{415, "bad content type"}
		}

		var (
			err     error
			buffer  []byte
			message Message
		)

		if buffer, err = ioutil.ReadAll(request.Body); err != nil {
			return &http_status{400, "invalid content stream"}
		}

		request.Body.Close()

		if json.Unmarshal(buffer, &message) != nil {
			return &http_status{400, "corrupt content format"}
		}

		err = server.listeners.Publish(int64(channel), &message)

		if err != nil {
			return &http_status{500, err.Error()}
		}

		writer.WriteHeader(200)
		fmt.Fprintln(writer, "Posted.")
	}

	return nil
}
