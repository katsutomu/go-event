package main

import "go-event/goevent"

func main() {
	go goevent.Subscribe("amqp://guest:guest@localhost:5672/", "test-queue", &goevent.TestHandler{})
	go goevent.Subscribe("amqp://guest:guest@localhost:5672/", "test-queue", &goevent.TestHandler{})
	select {}
}
