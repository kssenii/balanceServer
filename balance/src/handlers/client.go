package handlers

import (
	"context"
	"data"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Request struct {
	log *log.Logger
}

func NewRequest(l *log.Logger) *Request {
	return &Request{l}
}

func (request *Request) AddClient(rw http.ResponseWriter, r *http.Request) {

	request.log.Println("Creating a new request")
	newClient := r.Context().Value(KeyClient{}).(*data.Client)
	data.AddClient(newClient)
}

func (request *Request) GetClients(rw http.ResponseWriter, r *http.Request) {
	request.log.Println("Fetching all client's data")
	clientsList := data.GetCLients()
	err := clientsList.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to marshal json", http.StatusInternalServerError)
	}
}

func (request *Request) UpdateClientData(rw http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(rw, "Client ID is not convertable to uint64", http.StatusBadRequest)
		return
	}

	request.log.Printf("Creating client with ID %d", id)

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

func (request Request) MiddlewareValidateClient(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		currentClient := &data.Client{}

		err := currentClient.FromJSON(r.Body)
		if err != nil {
			request.log.Println("[ERROR] deserializing product", err)
			http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}

		// validate client's data
		err = currentClient.Validate()
		if err != nil {
			request.log.Println("[ERROR] validating parsed data", err)
			http.Error(rw, "Error validating data", http.StatusBadRequest)
			return
		}

		// add the product to the context
		context := context.WithValue(r.Context(), KeyClient{}, currentClient)
		r = r.WithContext(context)

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(rw, r)
	})
}
