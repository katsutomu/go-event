package goevent

import (
	"log"
	"time"
)

type Handler interface {
	Handle() error
}

type TestHandler struct {
	Test int64 `json:"test"`
}

func (t TestHandler) Handle() error {
	log.Printf("handle:%v\n", t.Test)
	time.Sleep(time.Second * 5)
	return nil
}
