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

func (request *Request) CheckBalance(id uint32) (uint64, error) {

	// Fetch balance with current id from the database
	balance, err := request.dbHandler.SelectData(id)
	if err != nil {
		request.log.Println("[ERROR] SELECT query failed. Reason: ", err)
		return 0, err
	}

	// if current id was not found in table, add this id with zero balance
	if len(balance) == 0 {

		request.log.Printf("[TRACE] Client with ID %d does not have a balance yet. Adding default data", id)

		err := request.dbHandler.InsertData(id)
		if err != nil {
			request.log.Printf("[ERROR] Unable to insert default data for ID %d. Reason: %s", id, err)
			return 0, err
		}
		balance = "0"
	}

	return strconv.ParseUint(balance, 10, 64)
}

func (request *Request) GetBalance(rw http.ResponseWriter, r *http.Request) {

	clientData := r.Context().Value(KeyClient{}).(*ClientBalanceData)

	// Fetch current balance if id is found in table, if not - add it to table with zero balance
	balance, err := request.CheckBalance(clientData.ToID)
	if err != nil {
		NotifyClient(rw, "Unable to get balance", http.StatusInternalServerError)
		return
	}
	clientData.Balance = balance

	// Send json data back to client
	err = clientData.ToJSON(rw)
	if err != nil {
		request.log.Println("[ERROR] Unable to parse JSON. Reason: ", err)
		NotifyClient(rw, "Unable to get balance", http.StatusInternalServerError)
		return
	}
}

func (request *Request) UpdateBalance(rw http.ResponseWriter, r *http.Request) {

	clientData := r.Context().Value(KeyClient{}).(*ClientBalanceData)

	// Fetch current balance if id is found in table, if not - add it to table with zero balance
	balance, err := request.CheckBalance(clientData.ToID)
	if err != nil {
		NotifyClient(rw, "Unable to get balance for client", http.StatusInternalServerError)
		return
	}
	clientData.Balance = balance

	// Calculate new balance for current id
	balanceDif := clientData.Sum
	if balanceDif >= 0 {
		clientData.Balance += uint64(balanceDif)
	} else {
		// make balance dif a positive number
		balanceDif *= -1

		// Do not allow to perform request if there is not enought balance
		if clientData.Balance >= uint64(balanceDif) {
			clientData.Balance -= uint64(balanceDif)
		} else {
			NotifyClient(rw, "Cannot satisfy request: not enough balance", http.StatusBadRequest)
			return
		}
	}

	// Update balance in database
	err = request.dbHandler.UpdateData(clientData.ToID, clientData.Balance)
	if err != nil {
		request.log.Printf("[ERROR] UPDATE query failed for ID %d. Reason: %s", clientData.ToID, err)
		NotifyClient(rw, "Unable to update client data", http.StatusInternalServerError)
		return
	}
}

func (request *Request) TransferBalance(rw http.ResponseWriter, r *http.Request) {

	data := r.Context().Value(KeyClient{}).(*ClientBalanceData)

	// Get current balance for clients, who receive money and who sends money
	receiverBalance, err1 := request.CheckBalance(data.ToID)
	senderBalance, err2 := request.CheckBalance(data.FromID)
	if err1 != nil || err2 != nil {
		NotifyClient(rw, "Unable to get balance for clients", http.StatusInternalServerError)
		return
	}

	// Sum to transfer should be a positive number, but if parsed value is negative - not a problem
	var sum uint32
	if data.Sum < 0 {
		sum = uint32(-1 * data.Sum)
	} else {
		sum = uint32(data.Sum)
	}

	// If clinet's balance, who sends money to another client, is not enought to satisfy the request, send an error
	if senderBalance < uint64(sum) {
		NotifyClient(rw, "Not enough balance to transfer money", http.StatusBadRequest)
		return
	}

	// Decrease balance of sender and increase balance of receiver
	err1 = request.dbHandler.UpdateData(data.ToID, receiverBalance+uint64(sum))
	err2 = request.dbHandler.UpdateData(data.FromID, senderBalance-uint64(sum))

	// If one of the balance updates returned an error, rollback both updates
	if err1 != nil || err2 != nil {
		if err1 == nil {
			request.dbHandler.UpdateData(data.ToID, receiverBalance)
			request.log.Printf("[ERROR] UPDATE query failed for ID %d. Reason: %s", data.FromID, err2)
		} else if err2 == nil {
			request.dbHandler.UpdateData(data.ToID, senderBalance)
			request.log.Printf("[ERROR] UPDATE query failed for ID %d. Reason: %s", data.ToID, err1)
		}

		NotifyClient(rw, "Unable to update client data", http.StatusInternalServerError)
	}
}

type KeyClient struct{}

func (request Request) MiddlewareValidateClient(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		currentClient := &ClientBalanceData{}

		// Parse client's data
		err := currentClient.FromJSON(r.Body)
		if err != nil {
			NotifyClient(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}

		// Validate client's data
		err = currentClient.Validate()
		if err != nil {
			NotifyClient(rw, "Unable to validate data", http.StatusBadRequest)
			return
		}

		// Add context
		context := context.WithValue(r.Context(), KeyClient{}, currentClient)
		r = r.WithContext(context)

		next.ServeHTTP(rw, r)
	})
}
