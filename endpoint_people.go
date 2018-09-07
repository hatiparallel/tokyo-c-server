package main

import (
	"net/http"
	"strings"

	"golang.org/x/net/context"
)

func endpoint_people(request *http.Request) *http_status {
	_, err := authenticate(request)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	parameter := strings.TrimPrefix(request.URL.Path, "/people/")

	record, err := idp.GetUser(context.Background(), parameter)

	if err != nil {
		return &http_status{500, "firebase failed: " + err.Error()}
	}

	return &http_status{200, record}
}
