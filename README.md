# Tokyo C Messanger Server

## Establish a Server

Execute `sudo go run server.go` and you have an endpoint at `http://localhost/messages`.

## Make a Listen
Just execute
`curl http://localhost/messages`.

## Publish a Message

JSON messages will be accepted:
`curl -X POST -H'Content-Type: application/json' -d '{"Text":"Hello, world!"}' http://localhost/messages`.
