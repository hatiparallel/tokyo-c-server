package main

import (
	"net/http"

	"golang.org/x/net/context"
)

func endpoint_people(request *http.Request) *http_status {
	_, err := authenticate(request)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	var person_id string

	err = match(request.URL.Path, "/people/([^/]+)", &person_id)

	if err != nil {
		return &http_status{400, err.Error()}
	}

	record, err := idp.GetUser(context.Background(), person_id)

	if err != nil {
		return &http_status{500, "firebase failed: " + err.Error()}
	}

	return &http_status{200, record}
}
