package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	var (
		port    int
		pidfile string
	)

	flag.IntVar(&port, "port", 80, "specifies port number to be binded")
	flag.StringVar(&pidfile, "pidfile", "/tmp/tokyo-c.pid", "specifies path to pidfile")

	flag.Parse()

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, "Hello, this is Tokyo C Server. It works!")
	})

	http.Handle("/stream/", NewMessageServer(func(message Message) error {
		posted_at := time.Now()

		message["PostedAt"] = posted_at

		if content, ok := message["Content"].(string); ok {
			fmt.Println(content)
		}

		return nil
	}))

	if file, err := os.OpenFile(pidfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755); err == nil {
		fmt.Fprintf(file, "%d\n", os.Getpid())
		file.Close()
	} else {
		fmt.Fprintf(os.Stderr, "failed to open pidfile\n")
		return
	}

	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}
