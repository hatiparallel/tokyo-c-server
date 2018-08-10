package main

import (
	"fmt"
	"net/http"
)

func main() {
	ch := make(chan string)

	http.HandleFunc("/messages", func (w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			header := w.Header()
			header.Set("Transfer-Encoding", "chunked")

			if f, ok := w.(http.Flusher); ok {
				for {
					fmt.Fprintf(w, "%s\n", <-ch)
					f.Flush()
				}
			}
		case "POST":
			var text string
			fmt.Fscan(r.Body, &text)
			ch <- text
			fmt.Fprintf(w, "Done.\n")
		}
	})

	http.ListenAndServe(":80", nil)
}
