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

	"golang.org/x/net/context"
	"firebase.google.com/go"
	firebase_auth "firebase.google.com/go/auth"
	_ "github.com/go-sql-driver/mysql"

	"google.golang.org/api/option"
)

type http_status struct {
	code    int
	message string
}

var db *sql.DB
var idp *firebase_auth.Client

func authenticate(request *http.Request, subject *string) error {
	var (
		auth_type string
		token     string
	)

	fmt.Sscanf(request.Header.Get("Authorization"), "%s %s", &auth_type, &token)

	if auth_type != "Bearer" {
		return errors.New("auth type must be Bearer")
	}

	verified, err := idp.VerifyIDToken(context.Background(), token)

	if err != nil {
		return errors.New("invalid token: "+err.Error())
	}

	*subject = verified.UID

	return nil
}

func main() {
	var (
		err error

		port    int
		pidfile string
		firebase_credentials string
	)

	flag.IntVar(&port, "port", 80, "specifies port number to be binded")
	flag.StringVar(&pidfile, "pidfile", "/tmp/tokyo-c.pid", "specifies path to pidfile")
	flag.StringVar(&firebase_credentials, "firebase", "firebase-credentials.json", "specifies path to firebase credentials")

	flag.Parse()

	db, err = sql.Open(os.Getenv("DATABASE_TYPE"), os.Getenv("DATABASE_URI")+"?parseTime=true")

	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to connect database: "+err.Error())
		return
	}

	firebase_app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(firebase_credentials))

	if err != nil {
		fmt.Fprintln(os.Stderr, "firebase initialization failed: "+err.Error())
		return
	}

	idp, err = firebase_app.Auth(context.Background())

	if err != nil {
		fmt.Fprintln(os.Stderr, "firebase authentication failed: "+err.Error())
		return
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, "Hello, this is Tokyo C Server. It works!")
	})

	http.HandleFunc("/friendships/", func(writer http.ResponseWriter, request *http.Request) {
		if status := endpoint_friendships(writer, request); status != nil {
			writer.WriteHeader(status.code)
			fmt.Fprintln(writer, status.message)
		}
	})

	http.HandleFunc("/channels/", func(writer http.ResponseWriter, request *http.Request) {
		if status := endpoint_channels(writer, request); status != nil {
			writer.WriteHeader(status.code)
			fmt.Fprintln(writer, status.message)
		}
	})

	message_handler := NewMessageServer(stamp_message)

	http.HandleFunc("/streams/", func(writer http.ResponseWriter, request *http.Request) {
		var subject string

		err := authenticate(request, &subject)

		if err != nil {
			writer.WriteHeader(401)
			fmt.Fprintln(writer, err.Error())
			return
		}

		if status := message_handler.handle_request(writer, request); status != nil {
			writer.WriteHeader(status.code)
			fmt.Fprintln(writer, status.message)
		}
	})

	http.HandleFunc("/messages/", func(writer http.ResponseWriter, request *http.Request) {
		if status := endpoint_messages(writer, request); status != nil {
			writer.WriteHeader(status.code)
			fmt.Fprintln(writer, status.message)
		}
	})

	http.HandleFunc("/people/", func(writer http.ResponseWriter, request *http.Request) {
		if status := endpoint_people(writer, request); status != nil {
			writer.WriteHeader(status.code)
			fmt.Fprintln(writer, status.message)
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

func stamp_message(channel_id int64, message *Message) error {
	message.Channel = channel_id
	message.PostedAt = time.Now()

	result, err := db.Exec(
		"INSERT INTO messages (channel, author, is_event, posted_at, content) VALUES (?, ?, ?, ?, ?)",
		channel_id, 0, 0, message.PostedAt, message.Content)

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
