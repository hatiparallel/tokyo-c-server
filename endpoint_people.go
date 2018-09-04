package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/net/context"
)

func endpoint_people(writer http.ResponseWriter, request *http.Request) *http_status {
	var subject string

	err := authenticate(request, &subject)

	if err != nil {
		return &http_status{401, err.Error()}
	}

	parameter := strings.TrimPrefix(request.URL.Path, "/people/")

	record, err := idp.GetUser(context.Background(), parameter)

	if err != nil {
		return &http_status{500, "firebase failed: " + err.Error()}
	}

	buffer, err := json.Marshal(record)

	if err != nil {
		return &http_status{500, err.Error()}
	}

	writer.Write(buffer)

	return nil
}
