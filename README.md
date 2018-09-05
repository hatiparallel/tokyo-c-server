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
`curl http://localhost:9000/streams/12345`.

## Publish a Message

JSON messages will be accepted:
`curl -X POST -H'Authorization: Bearer token' -H'Content-Type: application/json' -d '{"Content":"Hello, world!"}' http://localhost:9000/streams/12345`.

## Endpoints

* **GET** /streams/_channel_ gives a realtime stream of messages of _channel_.
	* Response (application/json stream): Message (data_types.go)
* **POST** /streams/_channel_ takes a message and broadcast it on _channel_
   * Payload (application/json): {IsEvent int; Content string}
* **GET** /messages/_channel_?since_id=_id_ gives a pile of messages of _channel_ since _id_.
	* Response (application/json): []Message (data_types.go)
* **GET** /friendships/ gives a friend list of the current user.
	* Repsponse (application/json): []string
* **POST** /friendships/ takes PIN and sends a request to the owner.
	* Payload (text/plain): PIN code
* **PUT** /friendships/_person_ makes a friendship from the current user to _person_ (requiring that _person_ send a request with PIN).
* **DELETE** /friendships/_person_ dissolves a friendship from the current user to _person_.
* **GET** /channels/ gives a channel list the current user is participating in.
	* Response (application/json): {Name string; Members []string}
* **POST** /channels/ makes a new channel with the only participant being the current user.
   * Payload (application/json): {Name string; Members [] string}
* **PUT** /channels/_channel_/_person_ makes _person_ participate in _channel_ (requiring that the current user be a member).
* **DELETE** /channels/_channel_/_person_ makes _person_ withdraw from _channel_. (requiring that the current user be a member. the channel will perish if the current user is the last participant).
* **GET** /pin gives a realtime stream of event messages
	* Response (application/json stream): {Type string; PIN int; Person string} (Type is `pin` -> `request` or `noop`)
