package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type StartupState struct {
	log *log.Logger
}

func Startup(l *log.Logger) *StartupState {
	return &StartupState{l}
}

func (data *StartupState) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	data.log.Println("Starting up")
	input, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(rw, "Error", http.StatusBadRequest)
		return
	}

	data.log.Printf("Client %s", input)
	fmt.Fprintf(rw, "Hi %s", input)
}
