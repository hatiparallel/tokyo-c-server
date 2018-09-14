# Tokyo C Messenger Server

## Build

`dep ensure` and `go build -o ~/bin/tokyo-c-server` to build into `~/bin/tokyo-c-server`.


## Make a Start Script

`cat > start-server` below and `chmod +x start-server`.

```
#!/bin/bash

export TOKYO_C_DATABASE_HOST="          "
export TOKYO_C_DATABASE_USER="          "
export TOKYO_C_DATABASE_PASSWORD="      "
export TOKYO_C_DATABASE_NAME="tokyoC_`date +%Y%m%d%H%M%S`"
export TOKYO_C_DATABASE_URI="${TOKYO_C_DATABASE_USER}:${TOKYO_C_DATABASE_PASSWORD}@tcp(${TOKYO_C_DATABASE_HOST})/${TOKYO_C_DATABASE_NAME}"

./install-database

DATABASE_TYPE=mysql DATABASE_URI="${TOKYO_C_DATABASE_URI}" go run *.go -port 9000 &

PID="$!"

mysql --user="${TOKYO_C_DATABASE_USER}" --host="${TOKYO_C_DATABASE_HOST}" --password="${TOKYO_C_DATABASE_PASSWORD}" "${TOKYO_C_DATABASE_NAME}"

kill -KILL "${PID}"

mysql --user="${TOKYO_C_DATABASE_USER}" --host="${TOKYO_C_DATABASE_HOST}" --password="${TOKYO_C_DATABASE_PASSWORD}" <<EOS
DROP DATABASE ${TOKYO_C_DATABASE_NAME};
EOS
```

## Place a Firebase Credentials

`cat > firebase-credentials.json`

## Establish a Server

Execute `./start-server`

## Endpoints

### /messages

* **GET** /messages gives a realtime stream of messages of _channel_.
	* Parameters
   		* `channel` specifies a channel which you want to listen.
		* `since_id` (optional) specifies a message id from which digging message archive starts.
	* Response (application/json stream): \[\][Message (data_types.go)](https://github.com/line-school2018summer/tokyo-c-server/blob/6dbe5771233705e67d86721610ddffbf732424d3/data_types.go#L33)
* **POST** /messages takes a message and broadcast it.
	* Payload (application/json): [Message (data_types.go)](https://github.com/line-school2018summer/tokyo-c-server/blob/6dbe5771233705e67d86721610ddffbf732424d3/data_types.go#L33)

### /messages/*

* **GET** /messages/_id_ gives a message of _id_.
	* Response (application/json stream): [Message (data_types.go)](https://github.com/line-school2018summer/tokyo-c-server/blob/6dbe5771233705e67d86721610ddffbf732424d3/data_types.go#L33)

### /friendships

* **GET** /friendships gives a friend list of the current user.
	* Repsponse (application/json): []string
* **POST** /friendships takes PIN and sends a request to the owner.
	* Payload (application/json): PIN code

### /friendships/*

* **PUT** /friendships/_person_ makes a friendship from the current user to _person_ (requiring that _person_ send a request with PIN).
	* Repsponse (application/json): []string
* **DELETE** /friendships/_person_ dissolves a friendship from the current user to _person_.
	* Repsponse (application/json): []string

### /channels

* **GET** /channels gives a channel list the current user is participating in.
	* Response (application/json): [Channel (data_type.go)](https://github.com/line-school2018summer/tokyo-c-server/blob/e43d2eeea6eb0270ec11d93c037d159f6ab837da/data_types.go#L27)
* **POST** /channels makes a new channel with the only participant being the current user.
   * Payload (application/json): [Channel (data_type.go)](https://github.com/line-school2018summer/tokyo-c-server/blob/e43d2eeea6eb0270ec11d93c037d159f6ab837da/data_types.go#L27)

### /channels/*

* **POST** /channels/_channel_/_person_ performs a bulk invitation.
	* Payload (application/json): []string
	* Response (application/json): [Channel (data_type.go)](https://github.com/line-school2018summer/tokyo-c-server/blob/e43d2eeea6eb0270ec11d93c037d159f6ab837da/data_types.go#L27)
* **PUT** /channels/_channel_/_person_ makes _person_ participate in _channel_ (requiring that the current user be a member).
	* Response (application/json): [Channel (data_type.go)](https://github.com/line-school2018summer/tokyo-c-server/blob/e43d2eeea6eb0270ec11d93c037d159f6ab837da/data_types.go#L27)
* **DELETE** /channels/_channel_/_person_ makes _person_ withdraw from _channel_. (requiring that the current user be a member. the channel will perish if the current user is the last participant).
	* Response (application/json): [Channel (data_type.go)](https://github.com/line-school2018summer/tokyo-c-server/blob/e43d2eeea6eb0270ec11d93c037d159f6ab837da/data_types.go#L27)
* **PATCH** /channels/_channel_ takes a differential data and modifies the information of _channel_.
	* Payload (application/x-www-form-urlencoded): name (optoinal)

### /pin

* **GET** /pin gives a realtime stream of event messages
	* Response (application/json stream): [PINEvent (data_types.go)](https://github.com/line-school2018summer/tokyo-c-server/blob/e43d2eeea6eb0270ec11d93c037d159f6ab837da/data_types.go#L21) (Type is `pin` -> `request` or `noop`)

### /people/*

* **GET** /people/_person_ gives a firebase user info of _person_
	* Response (application/json): Refer to [auth.UserInfo (firebase)](https://godoc.org/firebase.google.com/go/auth#UserInfo)

### /status

* **GET** /status gives the current status of the current user
	* Response (application/json): [Status (data_types.go)](https://github.com/line-school2018summer/tokyo-c-server/blob/e43d2eeea6eb0270ec11d93c037d159f6ab837da/data_types.go#L48)
