package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func authenticate(request *http.Request) error {
	var (
		auth_type string
		token string
	)

	fmt.Sscanf(request.Header.Get("Authorization"), "%s %s", &auth_type, &token)

	if auth_type != "Bearer" {
		return errors.New("auth type must be Bearer")
	}

	return nil
}

func main() {
	var (
		err error

		port    int
		pidfile string
	)

	flag.IntVar(&port, "port", 80, "specifies port number to be binded")
	flag.StringVar(&pidfile, "pidfile", "/tmp/tokyo-c.pid", "specifies path to pidfile")

	flag.Parse()

	message_handler := NewMessageServer(handle_message)

	db, err = sql.Open(os.Getenv("DATABASE_TYPE"), os.Getenv("DATABASE_URI") + "?parseTime=true")

	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to connect database; check environment variables DATABASE_TYPE and DATABASE_URI")
		return
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, "Hello, this is Tokyo C Server. It works!")
	})

	http.HandleFunc("/friendships", func(writer http.ResponseWriter, request *http.Request) {
		if authenticate(request) != nil {
			writer.WriteHeader(401)
			fmt.Fprintln(writer, "auth failed")
			return
		}

		// TODO
	})

	http.HandleFunc("/streams/", func(writer http.ResponseWriter, request *http.Request) {
		if authenticate(request) != nil {
			writer.WriteHeader(401)
			fmt.Fprintln(writer, "auth failed")
			return
		}

		message_handler.ServeHTTP(writer, request)
	})

	http.HandleFunc("/messages/", func(writer http.ResponseWriter, request *http.Request) {
		if authenticate(request) != nil {
			writer.WriteHeader(401)
			fmt.Fprintln(writer, "auth failed")
			return
		}

		if request.Method != "GET" {
			writer.WriteHeader(405)
			fmt.Println(writer, "method not allowed")
		}

		channel, err := strconv.Atoi(strings.TrimPrefix(request.URL.Path, "/messages/"))

		if err != nil {
			writer.WriteHeader(400)
			fmt.Fprintln(writer, "invalid channel name")
			return
		}

		since_id, err := strconv.Atoi(request.FormValue("since_id"))

		if err != nil {
			writer.WriteHeader(400)
			fmt.Fprintln(writer, "invalid since_id")
			return
		}

		rows, err := db.Query("SELECT id, channel, author, is_event, posted_at, content FROM messages WHERE channel = ? AND id > ?", channel, since_id)

		if err != nil {
			writer.WriteHeader(500)
			fmt.Fprintln(writer, "internal server error: " + err.Error())
			return
		}

		defer rows.Close()

		var message Message

		for rows.Next() {
			err := rows.Scan(
				&message.Id,
				&message.Channel,
				&message.Author,
				&message.IsEvent,
				&message.PostedAt,
				&message.Content)
			
			if err != nil {
				break
			}

			buffer, err := json.Marshal(message)

			if err != nil {
				break
			}

			fmt.Fprintln(writer, string(buffer))
		}
	})

	if file, err := os.OpenFile(pidfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755); err == nil {
		fmt.Fprintf(file, "%d\n", os.Getpid())
		file.Close()
	} else {
		fmt.Fprintln(os.Stderr, "failed to open pidfile")
		return
	}


	fmt.Fprintln(os.Stderr, "Start listening....")

	if http.ListenAndServe(":"+strconv.Itoa(port), nil) != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
}

func handle_message(channel Channel, message Message) error {
	message.Channel = channel
	message.PostedAt = time.Now()

	result, err := db.Exec(
		"INSERT INTO messages (channel, author, is_event, posted_at, content) VALUES (?, ?, ?, ?, ?)",
		channel, 0, 0, message.PostedAt, message.Content)
	
	if err != nil {
		return errors.New("failed to store message because... " + err.Error())
	}

	id, err := result.LastInsertId()

	if err != nil {
		return errors.New("failed to get message id")
	}

	message.Id = id

	return nil
}

