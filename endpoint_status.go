package main

import (
	"net/http"
)

func endpoint_status(request *http.Request) *http_status {
	subject, err := authenticate(request)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	rows, err := db.Query("SELECT memberships.channel, MAX(id) FROM memberships, messages WHERE person = ? AND memberships.channel = messages.channel GROUP BY memberships.channel", subject)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	var status Status

	status.Latests = make(map[int]int)

	for rows.Next() {
		var channel, latest_id int

		if err := rows.Scan(&channel, &latest_id); err != nil {
			return &http_status{500, err.Error()}
		}

		status.Latests[channel] = latest_id
	}

	row := db.QueryRow("SELECT COUNT(*) FROM friendships WHERE person_0 = ? OR person_1 = ?", subject, subject)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	if err := row.Scan(&status.FriendshipCount); err != nil {
		return &http_status{500, err.Error()}
	}

	return &http_status{200, status}
}
