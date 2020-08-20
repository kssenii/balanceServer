package handlers

import (
	"log"
	"net/http"
)

type ShutdownState struct {
	log *log.Logger
}

func Shutdown(l *log.Logger) *ShutdownState {
	return &ShutdownState{l}
}

func (data *ShutdownState) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	data.log.Println("Shutting down")
	rw.Write([]byte("Shutting down"))
}
