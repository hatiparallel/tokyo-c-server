# Tokyo C Messenger Server

## Build

`dep ensure` and `go build -o ~/bin/tokyo-c-server` to build into `~/bin/tokyo-c-server`.


## Setup a Database

The credential storage `.env` should look lile

```
TOKYO_C_DATABASE_HOST="127.0.0.1"
TOKYO_C_DATABASE_USER="root"
TOKYO_C_DATABASE_PASSWORD="PASSWORD"
TOKYO_C_DATABASE_NAME="tokyoC_DB"
```

and execute `./install-database .env`.

## Place a Firebase Credentials

`cat > firebase-credentials.json`

## Establish a Server

Execute `DATABASE_TYPE=mysql DATABASE_URI="root:PASSWORD@tcp(127.0.0.1)/tokyoC_DB" ~/bin/tokyo-c-server -port 9000` and you have an endpoint at `http://localhost:9000/`.

## Make a Listen
Just execute
`curl http://localhost:9000/messages/12345`.

## Publish a Message

JSON messages will be accepted:
`curl -X POST -H'Authorization: Bearer token' -H'Content-Type: application/json' -d '{"Content":"Hello, world!"}' http://localhost:9000/messages/12345`.

## Endpoints

* **GET** /messages/_channel_ gives a realtime stream of messages of _channel_.
	* Parameters: `since_id` specifies a message id from which digging message archive starts.
	* Response (application/json stream): [\[\]Message (data_types.go)](https://github.com/line-school2018summer/tokyo-c-server/blob/973bfbc6a111abb311bbe61610e4d93e16471779/data_types.go#L33)
* **POST** /messages/_channel_ takes a message and broadcast it on _channel_
   * Payload (application/json): [Message (data_types.go)](https://github.com/line-school2018summer/tokyo-c-server/blob/973bfbc6a111abb311bbe61610e4d93e16471779/data_types.go#L33)
* **GET** /friendships/ gives a friend list of the current user.
	* Repsponse (application/json): []string
* **POST** /friendships/ takes PIN and sends a request to the owner.
	* Payload (application/json): PIN code
* **PUT** /friendships/_person_ makes a friendship from the current user to _person_ (requiring that _person_ send a request with PIN).
	* Repsponse (application/json): []string
* **DELETE** /friendships/_person_ dissolves a friendship from the current user to _person_.
	* Repsponse (application/json): []string
* **GET** /channels/ gives a channel list the current user is participating in.
	* Response (application/json): [Channel (data_type.go)](https://github.com/line-school2018summer/tokyo-c-server/blob/973bfbc6a111abb311bbe61610e4d93e16471779/data_types.go#L27)
* **POST** /channels/ makes a new channel with the only participant being the current user.
   * Payload (application/json): [Channel (data_type.go)](https://github.com/line-school2018summer/tokyo-c-server/blob/973bfbc6a111abb311bbe61610e4d93e16471779/data_types.go#L27)
* **PUT** /channels/_channel_/_person_ makes _person_ participate in _channel_ (requiring that the current user be a member).
* **DELETE** /channels/_channel_/_person_ makes _person_ withdraw from _channel_. (requiring that the current user be a member. the channel will perish if the current user is the last participant).
* **GET** /pin gives a realtime stream of event messages
	* Response (application/json stream): [pin_event (data_types.go)](https://github.com/line-school2018summer/tokyo-c-server/blob/973bfbc6a111abb311bbe61610e4d93e16471779/data_types.go#L21) (Type is `pin` -> `request` or `noop`)
* **GET** /people/_person_ gives a firebase user info of _person_
	* Response (application/json): Refer to [auth.UserInfo (firebase)](https://godoc.org/firebase.google.com/go/auth#UserInfo)
* **GET** /status gives the current status of the current user
	* Response (application/json): [Status (data_types.go)](https://github.com/line-school2018summer/tokyo-c-server/blob/973bfbc6a111abb311bbe61610e4d93e16471779/data_types.go#L42)
