package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func endpoint_messages(request *http.Request) *http_status {
	subject, err := authenticate(request)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	channel_id, err := strconv.Atoi(strings.TrimPrefix(request.URL.Path, "/messages/"))

	if err != nil {
		return &http_status{400, "invalid channel id"}
	}

	switch request.Method {
	case "GET":
		since_id, _ := strconv.Atoi(request.FormValue("since_id"))

		if since_id > 0 {
			rows, err := db.Query("SELECT id, channel, author, is_event, posted_at, content FROM messages WHERE channel = ? AND id > ?", channel_id, since_id)

			if err != nil {
				return &http_status{500, err.Error()}
			}

			return &http_status{200, func(encoder *json.Encoder) {
				var message Message

				for rows.Next() {
					if rows.Scan(&message.Id, &message.Channel, &message.Author, &message.IsEvent, &message.PostedAt, &message.Content) != nil {
						break
					}

					encoder.Encode(message)
				}

				rows.Close()

				listener := make(chan Message)

				hub.Subscribe(channel_id, listener)
				defer hub.Unsubscribe(listener)

				for {
					if encoder.Encode(<-listener) != nil {
						break
					}
				}
			}}
		} else {
			return &http_status{200, func(encoder *json.Encoder) {
				listener := make(chan Message)

				hub.Subscribe(channel_id, listener)
				defer hub.Unsubscribe(listener)

				for {
					if encoder.Encode(<-listener) != nil {
						break
					}
				}
			}}
		}
	case "POST":
		var message Message

		err := decode_payload(request, &message)

		if err != nil {
			return &http_status{400, err.Error()}
		}

		message.Channel = channel_id
		message.Author = subject

		err = stamp_message(&message)

		if err != nil {
			return &http_status{500, "stamp failed: " + err.Error()}
		}

		hub.Publish(channel_id, message)

		return &http_status{200, message}
	default:
		return &http_status{405, "method not allowed"}
	}
}
