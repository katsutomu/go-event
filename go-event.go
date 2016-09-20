package main

import "go-event/goevent"

func main() {
	goevent.Subscribe("amqp://guest:guest@localhost:5672/", "test-queue", &goevent.TestHandler{})
}
