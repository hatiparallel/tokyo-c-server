package main

import (
	"strconv"
	"encoding/json"
	"strings"
	"net/http"
	"io/ioutil"
)

func endpoint_friendships(writer http.ResponseWriter, request *http.Request) *http_status {
	var subject string

	err := authenticate(request, &subject)

	if err!= nil {
		return &http_status{401, err.Error()}
	}

	person_id := strings.TrimPrefix(request.URL.Path, "/friendships/")

	if person_id == "" {
		return endpoint_friendships_without_person_id(subject, writer, request)
	}

	switch request.Method {
	case "PUT":
		pin_table.mutex.Lock()
		ticket := pin_table.store[pin_table.inverse[subject]]
		ticket.mutex.Lock()
		defer ticket.mutex.Unlock()
		pin_table.mutex.Unlock()

		if _, exists := ticket.pendings[person_id]; !exists {
			return &http_status{403, "unable to approve unsent request"}
		}

		delete(ticket.pendings, person_id)

		if _, err := db.Exec("INSERT INTO friendships (person_0, person_1) VALUES (?, ?)", subject, person_id); err != nil {
			return &http_status{500, err.Error()}
		}
	case "DELETE":
		if _, err := db.Exec("DELETE FROM friendships WHERE (person_0 = $1 AND person_1 = $2) OR (person_0 = $2 AND person_1 = $1)", subject, person_id); err != nil {
			return &http_status{500, err.Error()}
		}
	default:
		return &http_status{405, "method not allowed"}
	}

	return write_friendships(subject, writer)
}

func endpoint_friendships_without_person_id(subject string, writer http.ResponseWriter, request *http.Request) *http_status {
	switch request.Method {
	case "GET":
		return write_friendships(subject, writer)
	case "POST":
		if request.Header.Get("Content-Type") != "text/plain" {
			return &http_status{415, "bad content type"}
		}

		buffer, err := ioutil.ReadAll(request.Body)

		if err != nil {
			return &http_status{400, "invalid content stream"}
		}

		request.Body.Close()

		pin, err := strconv.Atoi(string(buffer))

		if err != nil {
			return &http_status{400, "corrupt content format"}
		}

		pin_table.mutex.Lock()

		ticket, pin_exists := pin_table.store[pin]
		ticket.mutex.Lock()
		defer ticket.mutex.Unlock()
		pin_table.mutex.Unlock()

		if !pin_exists {
			return &http_status{400, "nonexistent PIN"}
		}

		if subject == ticket.owner {
			return &http_status{400, "oneself cannot be his friend"}
		}

		if _, exists := ticket.pendings[subject]; exists {
			return &http_status{400, "you already sent request"}
		}

		ticket.channel <- subject
		ticket.pendings[subject] = true

		return &http_status{201, "wait for your request approved"}
	default:
		return &http_status{405, "method not allowed"}
	}
}

func write_friendships(subject string, writer http.ResponseWriter) *http_status {
	rows, err := db.Query("SELECT person_0, person_1 FROM friendships WHERE person_0 = $1 OR person_1 = $1", subject)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	friends := make([]string, 0, 16)

	var person_0, person_1 string

	for rows.Next() {
		if err := rows.Scan(&person_0, &person_1); err != nil {
			return &http_status{500, err.Error()}
		}

		if person_0 == subject {
			friends = append(friends, person_1)
		} else {
			friends = append(friends, person_0)
		}
	}

	buffer, err := json.Marshal(friends)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(buffer)

	return nil
}
