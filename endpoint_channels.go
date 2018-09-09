package main

import (
	"net/http"
	"strconv"
	"strings"
)

func endpoint_channels(request *http.Request) *http_status {
	subject, err := authenticate(request)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	parameter := strings.TrimPrefix(request.URL.Path, "/channels/")

	if parameter == "" {
		return endpoint_channels_without_parameter(subject, request)
	}

	parameter_splited := strings.SplitN(parameter, "/", 2)

	if len(parameter_splited) < 2 {
		parameter_splited = append(parameter_splited, "")
	}

	channel_id, err := strconv.Atoi(parameter_splited[0])

	if err != nil {
		return &http_status{400, "failed to parse channel id"}
	}

	person_id := parameter_splited[1]

	if db.QueryRow("SELECT person FROM memberships WHERE person = ?", subject).Scan(&subject) != nil {
		return &http_status{403, "not a member"}
	}

	switch request.Method {
	case "GET":
		if person_id != "" {
			return &http_status{400, "person parameter cannot be specified"}
		}
	case "PUT":
		if _, err := db.Exec("INSERT INTO memberships (channel, person) VALUES (?, ?)", channel_id, person_id); err != nil {
			return &http_status{500, err.Error()}
		}

		event := Message{
			Channel: channel_id,
			Author:  person_id,
			IsEvent: 1,
			Content: "join",
		}

		stamp_message(&event)
		hub.Publish(channel_id, event)
	case "DELETE":
		if _, err := db.Exec("DELETE FROM memberships WHERE channel = ? AND person = ?", channel_id, person_id); err != nil {
			return &http_status{500, err.Error()}
		}

		event := Message{
			Channel: channel_id,
			Author:  person_id,
			IsEvent: 1,
			Content: "leave",
		}

		stamp_message(&event)
		hub.Publish(channel_id, event)

		tx, err := db.Begin()

		if err != nil {
			return &http_status{500, err.Error()}
		}

		row := tx.QueryRow("SELECT COUNT(*) FROM memberships WHERE channel = ?", channel_id)

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

	var channel Channel

	if err := row.Scan(&channel.Name); err != nil {
		return &http_status{410, err.Error()}
	}

	rows, err := db.Query("SELECT person FROM memberships WHERE channel = ?", channel_id)
	channel.Members = make([]string, 0, 16)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	var person string

	for rows.Next() {
		if err := rows.Scan(&person); err != nil {
			return &http_status{500, err.Error()}
		}

		channel.Members = append(channel.Members, person)
	}

	return &http_status{200, channel}
}

func endpoint_channels_without_parameter(subject string, request *http.Request) *http_status {
	switch request.Method {
	case "GET":
		rows, err := db.Query("SELECT id, name FROM memberships, channels WHERE person = ? AND channel = id", subject)

		if err != nil {
			return &http_status{500, err.Error()}
		}

		var channel Channel

		channels := make([]Channel, 0, 16)

		for rows.Next() {
			if err := rows.Scan(&channel.Id, &channel.Name); err != nil {
				return &http_status{500, err.Error()}
			}

			channels = append(channels, channel)
		}

		return &http_status{200, channels}
	case "POST":
		tx, err := db.Begin()

		if err != nil {
			return &http_status{500, err.Error()}
		}

		var channel Channel

		err = decode_payload(request, &channel)

		if err != nil {
			return &http_status{400, err.Error()}
		}

		result, err := tx.Exec("INSERT INTO channels (name) VALUES (?)", channel.Name)

		if err != nil {
			tx.Rollback()
			return &http_status{500, err.Error()}
		}

		last_insert_id, err := result.LastInsertId()
		channel.Id = int(last_insert_id)

		if err != nil {
			tx.Rollback()
			return &http_status{500, err.Error()}
		}

		channel.Members = append(channel.Members, subject)

		for _, person_id := range channel.Members {
			if _, err = tx.Exec("INSERT INTO memberships (person, channel) VALUES (?, ?)", person_id, channel.Id); err != nil {
				tx.Rollback()
				return &http_status{500, err.Error()}
			}

			event := Message{
				Channel: channel.Id,
				Author:  person_id,
				IsEvent: 1,
				Content: "join",
			}

			stamp_message(&event)
			hub.Publish(channel.Id, event)
		}

		if err := tx.Commit(); err != nil {
			return &http_status{500, err.Error()}
		}

		return &http_status{200, channel}
	default:
		return &http_status{405, "method not allowed"}
	}
}
