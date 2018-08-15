# Tokyo C Messenger Server

## Build

`dep ensure` and `go build -o ~/bin/tokyo-c-server` to build into `~/bin/tokyo-c-server`.

## Establish a Server

Execute `~/bin/tokyo-c-server -port 9000` and you have an endpoint at `http://localhost:9000/`.

## Make a Listen
Just execute
`curl http://localhost:9000/messages/groupname`.

## Publish a Message

JSON messages will be accepted:
`curl -X POST -H'Content-Type: application/json' -d '{"Text":"Hello, world!"}' http://localhost:9000/messages/groupname`.
