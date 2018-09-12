package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func endpoint_messages(request *http.Request) *http_status {
	subject, err := authenticate(request)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	switch request.Method {
	case "GET":
		channel_id, err := strconv.Atoi(request.FormValue("channel"))

		if err != nil {
			return &http_status{400, err.Error()}
		}

		if db.QueryRow("SELECT person FROM memberships WHERE channel = ? AND person = ?", channel_id, subject).Scan(&subject) != nil {
			return &http_status{403, "not a member"}
		}

		return &http_status{200, func(encoder *json.Encoder) {
			var message Message

			since_id, err := strconv.Atoi(request.FormValue("since_id"))

			if err == nil {
				rows, err := db.Query("SELECT id, channel, author, is_event, posted_at, content FROM messages WHERE channel = ? AND id > ?", channel_id, since_id)

				if err == nil {
					for rows.Next() {
						if rows.Scan(&message.Id, &message.Channel, &message.Author, &message.IsEvent, &message.PostedAt, &message.Content) != nil {
							break
						}

						encoder.Encode(message)
					}
				}
			}

			listener := make(chan Message)

			hub.Subscribe(channel_id, listener)
			defer hub.Unsubscribe(listener)

			for {
				if encoder.Encode(<-listener) != nil {
					break
				}
			}
		}}
	case "POST":
		var message Message

		err := decode_payload(request, &message)

		if err != nil {
			return &http_status{400, err.Error()}
		}

		if db.QueryRow("SELECT person FROM memberships WHERE channel = ? AND person = ?", message.Channel, subject).Scan(&subject) != nil {
			return &http_status{403, "not a member"}
		}

		message.Author = subject

		err = stamp_message(&message)

		if err != nil {
			return &http_status{500, "stamp failed: " + err.Error()}
		}

		hub.Publish(message.Channel, message)

		return &http_status{200, message}
	default:
		return &http_status{405, "method not allowed"}
	}
}

func endpoint_messages_with_parameters(request *http.Request) *http_status {
	subject, err := authenticate(request)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	var message_id int

	err = match(request.URL.Path, "/messages/([0-9]+)", &message_id)

	if err != nil {
		return &http_status{400, err.Error()}
	}

	switch request.Method {
	case "GET":
		var message Message

		row := db.QueryRow(`
			SELECT id, channel, author, is_event, posted_at, content
			FROM messages
			WHERE id = ? AND channel IN (SELECT channel FROM memberships WHERE person = ?)`, message_id, subject)
		err := row.Scan(&message.Id, &message.Channel, &message.Author, &message.IsEvent, &message.PostedAt, &message.Content)

		if err != nil {
			return &http_status{500, err.Error()}
		}

		return &http_status{200, message}
	default:
		return &http_status{405, "method not allowed"}
	}
}
