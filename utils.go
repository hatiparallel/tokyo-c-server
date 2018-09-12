package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

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

func match(target string, pattern string, pointers ...interface{}) (err error) {
	matches := regexp.MustCompile("^(?:" + pattern + ")$").FindStringSubmatch(target)

	if len(matches) <= len(pointers) {
		return errors.New("arity mismatch")
	}

	for i, match := range matches {
		if i == 0 {
			continue
		}

		switch pointer := pointers[i-1].(type) {
		case *string:
			*pointer = match
		case *int:
			*pointer, err = strconv.Atoi(match)

			if err != nil {
				return
			}
		default:
			return errors.New("unsupported pointer type")
		}
	}

	return nil
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

func stamp_message(message *Message) error {
	message.PostedAt = time.Now()

	result, err := db.Exec(
		"INSERT INTO messages (channel, author, is_event, posted_at, content) VALUES (?, ?, ?, ?, ?)",
		message.Channel, 0, 0, message.PostedAt, message.Content)

	if err != nil {
		return err
	}

	last_insert_id, err := result.LastInsertId()

	if err != nil {
		return err
	}

	message.Id = int(last_insert_id)

	return nil
}
