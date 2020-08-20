package handlers

import (
	"data"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

type Client struct {
	log *log.Logger
}

func NewClient(l *log.Logger) *Client {
	return &Client{l}
}

func (client *Client) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		client.GetClients(rw, r)
		return
	}

	if r.Method == http.MethodPost {
		client.AddClient(rw, r)
		return
	}

	// expect ID in the URL
	if r.Method == http.MethodPut {
		regex := regexp.MustCompile(`/([0-9]+)`)
		g := regex.FindAllStringSubmatch(r.URL.Path, -1)
		if len(g) != 1 {
			client.log.Println("Invalid URI: more than one id", g)
			http.Error(rw, "Invalid URL", http.StatusBadRequest)
			return
		}
		if len(g[0]) != 2 {
			client.log.Println("Invalid URI: more than one capture group")
			http.Error(rw, "Invalid URL", http.StatusBadRequest)
			return
		}

		id, err := strconv.ParseUint(g[0][1], 10, 64)
		if err != nil {
			client.log.Println("Invalid URI: unable to convert to number")
			http.Error(rw, "Invalid URL", http.StatusBadRequest)
			return
		}

		client.log.Println("Registered client with id ", id)
		client.updateClientData(id, rw, r)

		return
	}

	rw.WriteHeader(http.StatusMethodNotAllowed)
}

func (client *Client) AddClient(rw http.ResponseWriter, r *http.Request) {
	client.log.Println("Creating a new client")

	newClient := &data.Client{}
	err := newClient.FromJSON(r.Body)
	if err != nil {
		http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
	}
	client.log.Printf("Client: %#v", newClient)
	data.AddClient(newClient)
}

func (client *Client) GetClients(rw http.ResponseWriter, r *http.Request) {
	client.log.Println("Fetching all client's data")
	clientsList := data.GetCLients()
	err := clientsList.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to marshal json", http.StatusInternalServerError)
	}
}

func (client *Client) updateClientData(id uint64, rw http.ResponseWriter, r *http.Request) {
	client.log.Printf("Creating client %d", id)

	currentClient := &data.Client{}
	err := currentClient.FromJSON(r.Body)
	if err != nil {
		http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
	}

	err = data.UpdateClient(id, currentClient)
	if err == data.ErrClientNotFound {
		http.Error(rw, "Client not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(rw, "Client not found", http.StatusInternalServerError)
		return
	}
}
