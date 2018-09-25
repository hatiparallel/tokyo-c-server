package main

import (
	"net/http"
)

func endpoint_channels(request *http.Request) *http_status {
	subject, err := authenticate(request)

	if err != nil {
		return &http_status{401, err.Error()}
	}

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
		}

		if err := tx.Commit(); err != nil {
			return &http_status{500, err.Error()}
		}

		for _, person_id := range channel.Members {
			event := Message{
				Channel: channel.Id,
				Author:  person_id,
				IsEvent: 1,
				Content: "join",
			}

			stamp_message(&event)
			hub.Publish(channel.Id, event)
		}

		return &http_status{200, channel}
	default:
		return &http_status{405, "method not allowed"}
	}
}

func endpoint_channels_with_parameters(request *http.Request) *http_status {
	subject, err := authenticate(request)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	var (
		channel_id int
		person_id  string
	)

	err = match(request.URL.Path, "/channels/([0-9]+)(?:/([^/]+))?", &channel_id, &person_id)

	if err != nil {
		return &http_status{400, err.Error()}
	}

	if db.QueryRow("SELECT person FROM memberships WHERE channel = ? AND person = ?", channel_id, subject).Scan(&subject) != nil {
		return &http_status{403, "not a member"}
	}

	switch request.Method {
	case "GET":
		if person_id != "" {
			return &http_status{400, "person parameter cannot be specified"}
		}
	case "POST":
		if person_id != "" {
			return &http_status{400, "person parameter cannot be specified"}
		}

		var friends []string

		decode_payload(request, &friends)

		tx, err := db.Begin()

		if err != nil {
			return &http_status{500, err.Error()}
		}

		for _, person_id := range friends {
			if _, err := db.Exec("INSERT INTO memberships (channel, person) VALUES (?, ?)", channel_id, person_id); err != nil {
				tx.Rollback()
				return &http_status{500, err.Error()}
			}
		}

		if err = tx.Commit(); err != nil {
			return &http_status{500, err.Error()}
		}

		for _, person_id := range friends {
			event := Message{
				Channel: channel_id,
				Author:  person_id,
				IsEvent: 1,
				Content: "join",
			}

			stamp_message(&event)
			hub.Publish(channel_id, event)
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

		if err = tx.Commit(); err != nil {
			return &http_status{500, err.Error()}
		}

		if size == 0 {
			return &http_status{410, "channel deleted"}
		}
	case "PATCH":
		request.ParseForm()


		if _, exists := request.PostForm["name"]; exists {
			_, err := db.Exec("UPDATE channels SET name = ? WHERE id = ?", request.PostForm.Get("name"), channel_id)

			if err != nil {
				return &http_status{500, err.Error()}
			}

			return &http_status{200, "channel name updated"}
		}

		return &http_status{200, "everything up to date"}
	default:
		return &http_status{405, "method not allowed"}
	}

	row := db.QueryRow("SELECT name FROM channels WHERE id = ?", channel_id)

	var channel Channel

	channel.Id = channel_id

	if err := row.Scan(&channel.Name); err != nil {
		return &http_status{500, err.Error()}
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
