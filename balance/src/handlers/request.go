package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Request struct {
	log       *log.Logger
	dbHandler *DBStorage
}

func NewRequest(dh *DBStorage) *Request {
	return &Request{
		log:       log.New(os.Stdout, "ClientLog ", log.LstdFlags),
		dbHandler: dh,
	}
}

func NotifyClient(rw http.ResponseWriter, errorMessage string, status int) {
	http.Error(rw, errorMessage, status)
}

func (request *Request) AddClient(rw http.ResponseWriter, r *http.Request) {

	data := r.Context().Value(KeyClient{}).(*ClientBalanceData)
	request.log.Printf("[TRACE] Adding client data with ID %d and zero balance", data.ID)

	err := request.dbHandler.InsertData(data.ID)
	if err != nil {
		request.log.Printf("[ERROR] Unable to insert default data for ID %d. Reason: %s", data.ID, err)
		NotifyClient(rw, "Unable to add client data", http.StatusInternalServerError)
	}
}

func (request *Request) GetBalance(rw http.ResponseWriter, r *http.Request) {

	clientData := r.Context().Value(KeyClient{}).(*ClientBalanceData)
	request.log.Println("[TRACE] Getting balance for client with ID ", clientData.ID)

	balance, err := request.dbHandler.SelectData(clientData.ID)
	if err != nil {
		request.log.Println("[ERROR] SELECT query failed. Reason: ", err)
		NotifyClient(rw, "Unable to get balance", http.StatusInternalServerError)
		return
	}

	clientData.Balance, _ = strconv.ParseUint(balance, 10, 64)
	err = clientData.ToJSON(rw)
	if err != nil {
		request.log.Println("[ERROR] Unable to parse JSON. Reason: ", err)
		NotifyClient(rw, "Unable to get balance", http.StatusInternalServerError)
	}
}

func (request *Request) UpdateBalance(rw http.ResponseWriter, r *http.Request) {

	clientData := r.Context().Value(KeyClient{}).(*ClientBalanceData)
	request.log.Printf("Updating balance %d for client with ID %d", clientData.Sum, clientData.ID)

	// err = UpdateClient(id, currentClient)
	// if err == ErrClientNotFound {
	// 	http.Error(rw, "Client not found", http.StatusNotFound)
	// 	return
	// }
	// if err != nil {
	// 	http.Error(rw, "Client not found", http.StatusInternalServerError)
	// 	return
	// }
	err := request.dbHandler.UpdateData(clientData.ID, clientData.Sum)
	if err != nil {
		request.log.Printf("[ERROR] Unable to update data for ID %d. Reason: %s", clientData.ID, err)
		NotifyClient(rw, "Unable to update client data", http.StatusInternalServerError)
	}
}

type KeyClient struct{}

func (request Request) MiddlewareValidateClient(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		currentClient := &ClientBalanceData{}

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

// func (request *Request) GetClientData(rw http.ResponseWriter, r *http.Request) {
// 	request.log.Println("Fetching all client's data")
// 	clientsList := GetCLients()
// 	err := clientsList.ToJSON(rw)
// 	if err != nil {
// 		http.Error(rw, "Unable to marshal json", http.StatusInternalServerError)
// 	}
// }
