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

and execute `./install-database`.

## Establish a Server

Execute `DATABASE_TYPE=mysql DATABASE_URI="root:PASSWORD@tcp(127.0.0.1)/tokyoC_DB" ~/bin/tokyo-c-server -port 9000` and you have an endpoint at `http://localhost:9000/`.

## Make a Listen
Just execute
`curl http://localhost:9000/stream/12345`.

## Publish a Message

JSON messages will be accepted:
`curl -X POST -H'Authorization: Bearer token' -H'Content-Type: application/json' -d '{"Content":"Hello, world!"}' http://localhost:9000/streams/12345`.

## Endpoints

* GET /streams/_channel_ gives a realtime stream of messages of _channel_.
* POST /streams/_channel_ takes a message and broadcast it on _channel_
   * Schema: {Id int64; Channel Channel; Author int64; IsEvent int; PostedAt time.Time; Content string}
* GET /messages/_channel_?since_id=_id_ gives a pile of messages of _channel_ since _id_.
