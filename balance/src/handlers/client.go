package handlers

import (
	"data"
	"log"
	"net/http"
)

type Client struct {
	log *log.Logger
}

func NewClient(l *log.Logger) *Client {
	return &Client{l}
}

func (client *Client) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	//data.log.Println("Client created")
	clientsList := data.GetCLients()
	err := clientsList.ToJSON(rw)
	//data, err := json.Marshal(clientsList)
	if err != nil {
		http.Error(rw, "Unable to marshal json", http.StatusInternalServerError)
	}

	//rw.Write(data)
}
