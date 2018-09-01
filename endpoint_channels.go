package main

import (
	"fmt"
	"encoding/json"
	"strconv"
	"strings"
	"net/http"
)

func endpoint_channels(writer http.ResponseWriter, request *http.Request) *http_status {
	var subject string

	err := authenticate(request, &subject)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	if err := request.ParseForm(); err != nil {
		return &http_status{401, "auth failed"}
	}

	parameter := strings.TrimPrefix(request.URL.Path, "/channels/")

	if request.Method == "GET" && parameter == "" {
		rows, err := db.Query("SELECT id, name FROM participations, channels WHERE person = ? AND channel = id", subject)

		var channel Channel

		channels := make([]Channel, 0, 16)

		for rows.Next() {
			if err := rows.Scan(&channel.Id, &channel.Name); err != nil {
				return &http_status{500, err.Error()}
			}

			channels = append(channels, channel)
		}

		buffer, err := json.Marshal(channels)

		if err != nil {
			return &http_status{500, err.Error()}
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(buffer)

		return nil
	}

	if request.Method == "POST" && parameter == "" {
		tx, err := db.Begin()

		if err != nil {
			return &http_status{500, err.Error()}
		}

		result, err := tx.Exec("INSERT INTO channels (name) VALUES (?)", request.PostForm.Get("name"))

		if err != nil {
			tx.Rollback()
			return &http_status{500, err.Error()}
		}

		channel_id, err := result.LastInsertId()

		if err != nil {
			tx.Rollback()
			return &http_status{500, err.Error()}
		}

		if _, err = tx.Exec("INSERT INTO participations (person, channel) VALUES (?, ?)", subject, channel_id); err != nil {
			tx.Rollback()
			return &http_status{500, err.Error()}
		}

		if err := tx.Commit(); err != nil {
			return &http_status{500, err.Error()}
		}

		fmt.Fprintf(writer, "%d", channel_id)

		return nil
	}

	channel_id, err := strconv.Atoi(parameter)

	if err != nil {
		return &http_status{400, "failed to parse channel id"}
	}

	switch request.Method {
	case "GET":

	case "POST":
		if _, err := db.Exec("INSERT INTO participations (channel, person) VALUES (?, ?)", channel_id, subject); err != nil {
			return &http_status{500, err.Error()}
		}
	case "DELETE":
		if _, err := db.Exec("DELETE FROM participations WHERE channel = ? AND person = ?", channel_id, subject); err != nil {
			return &http_status{500, err.Error()}
		}

		tx, err := db.Begin()

		if err != nil {
			return &http_status{500, err.Error()}
		}

		row := tx.QueryRow("SELECT COUNT(*) FROM participations WHERE channel = ?", channel_id)

		var size int

		if err := row.Scan(&size); err != nil {
			return &http_status{500, err.Error()}
		}

		if size == 0 {
			if _, err := tx.Exec("DELETE FROM channels WHERE id = ?", channel_id); err != nil {
				return &http_status{500, err.Error()}
			}
		}

		tx.Commit()
	default:
		return &http_status{405, "method not allowed"}
	}

	row := db.QueryRow("SELECT name FROM channels WHERE id = ?", channel_id)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	var channel struct {
		Name         string
		Participants []string
	}

	if err := row.Scan(&channel.Name); err != nil {
		return &http_status{410, err.Error()}
	}

	rows, err := db.Query("SELECT person FROM participations WHERE channel = ?", channel_id)
	channel.Participants = make([]string, 0, 16)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	var person string

	for rows.Next() {
		if err := rows.Scan(&person); err != nil {
			return &http_status{500, err.Error()}
		}

		channel.Participants = append(channel.Participants, person)
	}

	buffer, err := json.Marshal(channel)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(buffer)

	return nil
}
