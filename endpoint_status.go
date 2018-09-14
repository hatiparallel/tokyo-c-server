package main

import (
	"net/http"
	"github.com/go-sql-driver/mysql"
)

func endpoint_status(request *http.Request) *http_status {
	subject, err := authenticate(request)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	switch request.Method {
	case "GET":
		rows, err := db.Query(`
			SELECT channels.id, channels.name, MAX(messages.id)
			FROM channels, memberships, messages
			WHERE memberships.person = ? AND memberships.channel = channels.id AND messages.channel = channels.id
			GROUP BY memberships.channel`, subject)

		if err != nil {
			return &http_status{500, err.Error()}
		}

		var (
			status  Status
			summary Summary
		)

		status.Latests = make([]Summary, 0)

		for rows.Next() {
			if err := rows.Scan(&summary.ChannelId, &summary.ChannelName, &summary.MessageId); err != nil {
				return &http_status{500, err.Error()}
			}

			status.Latests = append(status.Latests, summary)
		}

		row := db.QueryRow("SELECT COUNT(*), MAX(created_at) FROM friendships WHERE person_0 = ? OR person_1 = ?", subject, subject)

		if err != nil {
			return &http_status{500, err.Error()}
		}

		var friendship_added_at mysql.NullTime

		if err := row.Scan(&status.FriendshipCount, &friendship_added_at); err != nil {
			return &http_status{500, err.Error()}
		}

		status.FriendshipAddedAt = friendship_added_at.Time

		return &http_status{200, status}
	default:
		return &http_status{405, "method not allowed"}
	}
}
