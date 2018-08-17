package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
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

	db, err = sql.Open(os.Getenv("DATABASE_TYPE"), os.Getenv("DATABASE_URI"))

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

	http.HandleFunc("/stream/", func(writer http.ResponseWriter, request *http.Request) {
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

		// TODO
	})

	if file, err := os.OpenFile(pidfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755); err == nil {
		fmt.Fprintf(file, "%d\n", os.Getpid())
		file.Close()
	} else {
		fmt.Fprintf(os.Stderr, "failed to open pidfile\n")
		return
	}


	fmt.Fprintln(os.Stderr, "Start listening....")

	if http.ListenAndServe(":"+strconv.Itoa(port), nil) != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
}

func handle_message(channel Channel, message Message) error {
	var (
		ok bool
		content string
	)

	posted_at := time.Now()

	message["posted_at"] = posted_at

	if content, ok = message["content"].(string); !ok {
		return errors.New("message must have `content' field")
	}

	if _, err := db.Exec(
		"INSERT INTO messages (channel, author, is_event, posted_at, content) VALUES (?, ?, ?, ?, ?)",
		channel, 0, 0, posted_at, content); err != nil {
		return errors.New("failed to store message because... " + err.Error())
	}

	return nil
}

