package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func endpoint_messages(writer http.ResponseWriter, request *http.Request) *http_status {
	var subject string

	err := authenticate(request, &subject)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	if request.Method != "GET" {
		return &http_status{405, "method not allowed"}
	}

	channel_id, err := strconv.Atoi(strings.TrimPrefix(request.URL.Path, "/messages/"))

	if err != nil {
		return &http_status{400, "invalid channel name"}
	}

	since_id, err := strconv.Atoi(request.FormValue("since_id"))

	if err != nil {
		return &http_status{400, "invalid since_id"}
	}

	rows, err := db.Query("SELECT id, channel, author, is_event, posted_at, content FROM messages WHERE channel = ? AND id > ?", channel_id, since_id)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	defer rows.Close()

	var message Message

	messages := make([]Message, 32)

	for rows.Next() {
		err := rows.Scan(&message.Id, &message.Channel, &message.Author, &message.IsEvent, &message.PostedAt, &message.Content)

		if err != nil {
			return &http_status{500, err.Error()}
		}

		messages = append(messages, message)
	}

	buffer, err := json.Marshal(message)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(buffer)

	return nil
}
