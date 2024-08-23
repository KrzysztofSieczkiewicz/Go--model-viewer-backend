package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

type Hello struct {
	loggger *log.Logger
}

func NewHello(logger *log.Logger) *Hello {
	return &Hello{logger}
}

func (h *Hello) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	h.loggger.Println("Hello world")

	data, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Ooooops", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(writer, "Hello %s\n", data)
}