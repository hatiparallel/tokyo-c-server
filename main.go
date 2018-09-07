package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"

	"firebase.google.com/go"
	firebase_auth "firebase.google.com/go/auth"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/context"

	"google.golang.org/api/option"
)

var pin_table struct {
	by_pin   map[int]*pin_ticket
	by_owner map[string]*pin_ticket
	mutex    *sync.Mutex
}

var hub *Hub
var db *sql.DB
var idp *firebase_auth.Client

func main() {
	var (
		err                  error
		port                 int
		pidfile              string
		firebase_credentials string
	)

	flag.IntVar(&port, "port", 80, "specifies port number to be binded")
	flag.StringVar(&pidfile, "pidfile", "/tmp/tokyo-c.pid", "specifies path to pidfile")
	flag.StringVar(&firebase_credentials, "firebase", "firebase-credentials.json", "specifies path to firebase credentials")

	flag.Parse()

	hub = NewHub()

	pin_table.by_pin = make(map[int]*pin_ticket)
	pin_table.by_owner = make(map[string]*pin_ticket)
	pin_table.mutex = new(sync.Mutex)

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

	endpoints := map[string]func(*http.Request) *http_status{
		"/": func(_ *http.Request) *http_status {
			return &http_status{200, "Hello, this is Tokyo C Server."}
		},
		"/friendships/": endpoint_friendships,
		"/channels/":    endpoint_channels,
		"/messages/":    endpoint_messages,
		"/people/":      endpoint_people,
		"/pin":          endpoint_pin,
		"/status":       endpoint_status,
	}

	for path, handler := range endpoints {
		http.HandleFunc(path, build_handler(handler))
	}

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

func build_handler(handler func(*http.Request) *http_status) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		status := handler(request)

		if status == nil {
			return
		}

		writer.Header().Set("Content-Type", "application/json")

		if start_stream, ok := status.body.(func(*json.Encoder)); ok {
			flusher, flushable := writer.(http.Flusher)

			if !flushable {
				writer.WriteHeader(400)
				writer.Write([]byte("streaming cannot be established"))
				return
			}

			writer.Header().Set("Transfer-Encoding", "chunked")
			writer.WriteHeader(status.code)
			start_stream(json.NewEncoder(&immediate_writer{writer, flusher}))
		} else {
			writer.WriteHeader(status.code)
			json.NewEncoder(writer).Encode(status.body)
		}
	}
}

func decode_payload(request *http.Request, pointer interface{}) error {
	if request.Header.Get("Content-Type") != "application/json" {
		return errors.New("bad content type")
	}

	buffer, err := ioutil.ReadAll(request.Body)

	if err != nil {
		return errors.New("invalid content stream: " + err.Error())
	}

	request.Body.Close()

	err = json.Unmarshal(buffer, pointer)

	if err != nil {
		return errors.New("corrupt content format: " + err.Error())
	}

	return nil
}

type immediate_writer struct {
	writer  io.Writer
	flusher http.Flusher
}

func (iw *immediate_writer) Write(b []byte) (n int, err error) {
	n, err = iw.writer.Write(b)
	iw.flusher.Flush()
	return
}
