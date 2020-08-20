package handlers

import (
	"context"
	"data"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Client struct {
	log *log.Logger
}

func NewClient(l *log.Logger) *Client {
	return &Client{l}
}

func (client *Client) AddClient(rw http.ResponseWriter, r *http.Request) {

	client.log.Println("Creating a new client")
	newClient := r.Context().Value(KeyClient{}).(*data.Client)
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

func (client *Client) UpdateClientData(rw http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(rw, "Client ID is not convertable to uint64", http.StatusBadRequest)
		return
	}

	client.log.Printf("Creating client with ID %d", id)

	currentClient := r.Context().Value(KeyClient{}).(*data.Client)

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

type KeyClient struct{}

func (client Client) MiddlewareValidateClient(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		currentClient := &data.Client{}

		err := currentClient.FromJSON(r.Body)
		if err != nil {
			client.log.Println("[ERROR] deserializing product", err)
			http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}

		// add the product to the context
		context := context.WithValue(r.Context(), KeyClient{}, currentClient)
		r = r.WithContext(context)

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(rw, r)
	})
}
