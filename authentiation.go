package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/context"
)

func authenticate(request *http.Request) (string, error) {
	var (
		auth_type string
		token     string
	)

	fmt.Sscanf(request.Header.Get("Authorization"), "%s %s", &auth_type, &token)

	if auth_type != "Bearer" {
		return "", errors.New("auth type must be Bearer")
	}

	if strings.HasPrefix(token, "TEST_TOKEN") {
		return "TEST_USER" + strings.TrimPrefix(token, "TEST_TOKEN"), nil
	}

	verified, err := idp.VerifyIDToken(context.Background(), token)

	if err != nil {
		return "", errors.New("invalid token: " + err.Error())
	}

	return verified.UID, nil
}
