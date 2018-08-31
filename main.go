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

func endpoint_friendships(writer http.ResponseWriter, request *http.Request) *http_status {
	var subject string

	err := authenticate(request, &subject)
	
	if err!= nil {
		return &http_status{401, err.Error()}
	}

	person_id, err := strconv.Atoi(strings.TrimPrefix(request.URL.Path, "/friendships/"))

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

func endpoint_channels(writer http.ResponseWriter, request *http.Request) *http_status {
	var subject string

	err := authenticate(request, &subject)
	
	if err != nil {
		return &http_status{401, err.Error()}
	}

	if err := request.ParseForm(); err != nil {
		return &http_status{401, "auth failed"}
	}

	parameter := strings.TrimPrefix(request.URL.Path, "/channels/")

	if request.Method == "GET" && parameter == "" {
		rows, err := db.Query("SELECT id, name FROM participations, channels WHERE person = ? AND channel = id", subject)

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
	}

	if request.Method == "POST" && parameter == "" {
		tx, err := db.Begin()

		if err != nil {
			return &http_status{500, err.Error()}
		}

		result, err := tx.Exec("INSERT INTO channels (name) VALUES (?)", request.PostForm.Get("name"))

		if err != nil {
			tx.Rollback()
			return &http_status{500, err.Error()}
		}

		channel_id, err := result.LastInsertId()

		if err != nil {
			tx.Rollback()
			return &http_status{500, err.Error()}
		}

		if _, err = tx.Exec("INSERT INTO participations (person, channel) VALUES (?, ?)", subject, channel_id); err != nil {
			tx.Rollback()
			return &http_status{500, err.Error()}
		}

		if err := tx.Commit(); err != nil {
			return &http_status{500, err.Error()}
		}

		fmt.Fprintf(writer, "%d", channel_id)

		return nil
	}

	channel_id, err := strconv.Atoi(parameter)

	if err != nil {
		return &http_status{400, "failed to parse channel id"}
	}

	switch request.Method {
	case "GET":

	case "POST":
		if _, err := db.Exec("INSERT INTO participations (channel, person) VALUES (?, ?)", channel_id, subject); err != nil {
			return &http_status{500, err.Error()}
		}
	case "DELETE":
		if _, err := db.Exec("DELETE FROM participations WHERE channel = ? AND person = ?", channel_id, subject); err != nil {
			return &http_status{500, err.Error()}
		}

		tx, err := db.Begin()

		if err != nil {
			return &http_status{500, err.Error()}
		}

		row := tx.QueryRow("SELECT COUNT(*) FROM participations WHERE channel = ?", channel_id)

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

	if err != nil {
		return &http_status{500, err.Error()}
	}

	var channel struct {
		Name         string
		Participants []string
	}

	if err := row.Scan(&channel.Name); err != nil {
		return &http_status{410, err.Error()}
	}

	rows, err := db.Query("SELECT person FROM participations WHERE channel = ?", channel_id)
	channel.Participants = make([]string, 0, 16)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	var person string

	for rows.Next() {
		if err := rows.Scan(&person); err != nil {
			return &http_status{500, err.Error()}
		}

		channel.Participants = append(channel.Participants, person)
	}

	buffer, err := json.Marshal(channel)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(buffer)

	return nil
}

func endpoint_messages(writer http.ResponseWriter, request *http.Request) *http_status {
	var subject string

	err := authenticate(request, &subject)
	
	if err != nil {
		return &http_status{401, err.Error()}
	}

	if request.Method != "GET" {
		return &http_status{405, "method not allowed"}
	}

	channel_id, err := strconv.Atoi(strings.TrimPrefix(request.URL.Path, "/messages/"))

	if err != nil {
		return &http_status{400, "invalid channel name"}
	}

	since_id, err := strconv.Atoi(request.FormValue("since_id"))

	if err != nil {
		return &http_status{400, "invalid since_id"}
	}

	rows, err := db.Query("SELECT id, channel, author, is_event, posted_at, content FROM messages WHERE channel = ? AND id > ?", channel_id, since_id)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	defer rows.Close()

	var message Message

	for rows.Next() {
		err := rows.Scan(&message.Id, &message.Channel, &message.Author, &message.IsEvent, &message.PostedAt, &message.Content)

		if err != nil {
			return &http_status{500, err.Error()}
		}

		buffer, err := json.Marshal(message)

		if err != nil {
			return &http_status{500, err.Error()}
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(buffer)
	}

	return nil
}
