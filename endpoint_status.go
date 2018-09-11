package main

import (
	"net/http"
)

func endpoint_status(request *http.Request) *http_status {
	subject, err := authenticate(request)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	rows, err := db.Query(`
		SELECT channel.id channel, channel.name, MAX(messages.id)
		FROM channels, memberships, messages
		WHERE memberships.person = ? AND memberships.channel = channel AND messages.channel = channel
		GROUP BY memberships.channel`, subject)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	var (
		status Status
		summary Summary
	)

	status.Latests = make([]Summary, 16)

	for rows.Next() {
		if err := rows.Scan(&summary.ChannelId, &summary.ChannelName, &summary.MessageId); err != nil {
			return &http_status{500, err.Error()}
		}

		status.Latests = append(status.Latests, summary)
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
