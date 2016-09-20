package goevent

import "fmt"

type Handler interface {
	Handle() error
}

type TestHandler struct {
	Test int64 `json:"test"`
}

func (t TestHandler) Handle() error {
	fmt.Println(t)
	return nil
}
