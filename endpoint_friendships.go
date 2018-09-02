package main

import (
	"encoding/json"
	"strings"
	"net/http"
)

func endpoint_friendships(writer http.ResponseWriter, request *http.Request) *http_status {
	var subject string

	err := authenticate(request, &subject)

	if err!= nil {
		return &http_status{401, err.Error()}
	}

	person_id := strings.TrimPrefix(request.URL.Path, "/friendships/")

	if request.Method != "GET" && err != nil {
		return &http_status{400, "failed to parse person id"}
	}

	switch request.Method {
	case "GET":

	case "PUT":
		if _, err := db.Exec("INSERT INTO friendships (person_0, person_1) VALUES (?, ?)", subject, person_id); err != nil {
			return &http_status{500, err.Error()}
		}
	case "DELETE":
		if _, err := db.Exec("DELETE FROM friendships WHERE person_0 = ? AND person_1 = ?", subject, person_id); err != nil {
			return &http_status{500, err.Error()}
		}
	default:
		return &http_status{405, "method not allowed"}
	}

	rows, err := db.Query("SELECT person_1 FROM friendships WHERE person_0 = ? AND person_1 = id", subject)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	friends := make([]string, 0, 16)

	var person string

	for rows.Next() {
		if err := rows.Scan(&person); err != nil {
			return &http_status{500, err.Error()}
		}

		friends = append(friends, person)
	}

	buffer, err := json.Marshal(friends)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(buffer)

	return nil
}
