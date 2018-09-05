package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func endpoint_channels(writer http.ResponseWriter, request *http.Request) *http_status {
	subject, err := authenticate(request)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	if err := request.ParseForm(); err != nil {
		return &http_status{401, "auth failed"}
	}

	parameter := strings.TrimPrefix(request.URL.Path, "/channels/")

	if parameter == "" {
		return endpoint_channels_without_parameter(subject, writer, request)
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

		message_server.listeners.Publish(int64(channel_id), &Message{IsEvent: 1, Content: "join"})
	case "DELETE":
		if _, err := db.Exec("DELETE FROM memberships WHERE channel = ? AND person = ?", channel_id, person_id); err != nil {
			return &http_status{500, err.Error()}
		}

		message_server.listeners.Publish(int64(channel_id), &Message{IsEvent: 1, Content: "leave"})

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

	var channel struct {
		Name    string
		Members []string
	}

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

	buffer, err := json.Marshal(channel)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(buffer)

	return nil
}

func endpoint_channels_without_parameter(subject string, writer http.ResponseWriter, request *http.Request) *http_status {
	switch request.Method {
	case "GET":
		rows, err := db.Query("SELECT id, name FROM memberships, channels WHERE person = ? AND channel = id", subject)

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
	case "POST":
		tx, err := db.Begin()

		if err != nil {
			return &http_status{500, err.Error()}
		}

		var channel_info struct {
			Name    string
			Members []string
		}

		if request.Header.Get("Content-Type") != "application/json" {
			return &http_status{415, "bad content type"}
		}

		buffer, err := ioutil.ReadAll(request.Body)

		if err != nil {
			return &http_status{400, "invalid content stream"}
		}

		request.Body.Close()

		if json.Unmarshal(buffer, &channel_info) != nil {
			return &http_status{400, "corrupt content format"}
		}

		result, err := tx.Exec("INSERT INTO channels (name) VALUES (?)", channel_info.Name)

		if err != nil {
			tx.Rollback()
			return &http_status{500, err.Error()}
		}

		channel_id, err := result.LastInsertId()

		if err != nil {
			tx.Rollback()
			return &http_status{500, err.Error()}
		}

		channel_info.Members = append(channel_info.Members, subject)

		for _, person_id := range channel_info.Members {
			if _, err = tx.Exec("INSERT INTO memberships (person, channel) VALUES (?, ?)", person_id, channel_id); err != nil {
				tx.Rollback()
				return &http_status{500, err.Error()}
			}
		}

		if err := tx.Commit(); err != nil {
			return &http_status{500, err.Error()}
		}

		message_server.listeners.Publish(int64(channel_id), &Message{IsEvent: 1, Content: "join"})

		fmt.Fprintf(writer, "%d", channel_id)

		return nil
	default:
		return &http_status{405, "method not allowed"}
	}
}
