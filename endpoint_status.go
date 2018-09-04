package main

import (
	"net/http"
	// "encoding/json"
)

func endpoint_status(writer http.ResponseWriter, request *http.Request) *http_status {
	var subject string

	err := authenticate(request, &subject)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	return nil
}
