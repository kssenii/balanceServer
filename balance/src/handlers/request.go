package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Request struct {
	log       *log.Logger
	dbHandler *DBStorage
}

func NewBalanceRequest(dh *DBStorage) *Request {
	return &Request{
		log:       log.New(os.Stdout, "ClientLog ", log.LstdFlags),
		dbHandler: dh,
	}
}

func NotifyClient(rw http.ResponseWriter, errorMessage string, status int) {
	http.Error(rw, errorMessage, status)
}

func (request *Request) CheckBalance(id uint32) (int32, error) {

	data := TableData{
		ClientID:  id,
		TableName: BALANCE_TABLE,
	}

	// Fetch balance with current id from the database
	selectedData, err := request.dbHandler.SelectData(data)
	if err != nil {
		request.log.Println("[ERROR] SELECT query failed. Reason: ", err)
		return 0, err
	}

	var balance int32

	// if current id was not found in table, add this id with zero balance
	if len(selectedData) == 0 {

		request.log.Printf("[TRACE] Client with ID %d does not have a balance yet. Adding default data", id)

		err := request.dbHandler.InsertData(data)
		if err != nil {
			request.log.Printf("[ERROR] Unable to insert default data for ID %d. Reason: %s", id, err)
			return 0, err
		}
		balance = 0
	} else {
		balance = selectedData[0].Sum
	}

	return balance, nil
}

func getConvertedBalance(currency string, amount int32) (float64, error) {

	URL := "https://api.exchangeratesapi.io/latest?base=USD&symbols=RUB"

	resp, err := http.Get(URL)
	if err != nil {
		log.Println("[ERROR] Request to get currency value failed", err)
		return 0, err
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	var rubOne float64
	rate := result["rates"].(map[string]interface{})
	if value, ok := rate["RUB"].(float64); ok {
		rubOne = value
	}
	convertedBalance := float64(amount) / rubOne

	return convertedBalance, nil
}

func (request *Request) LogCurrentTransaction(id uint32, sum int32, description string) error {

	err := request.dbHandler.InsertData(TableData{
		ClientID:    id,
		Sum:         sum,
		Description: description,
		TableName:   LOG_TABLE,
	})
	if err != nil {
		request.log.Printf("[ERROR] Unable to insert logging data for ID %d. Reason: %s", id, err)
		return err
	}

	return nil
}

func (request *Request) GetBalance(rw http.ResponseWriter, r *http.Request) {

	clientData := r.Context().Value(KeyClient{}).(*ClientBalanceData)

	// Fetch current balance if id is found in table, if not - add it to table with zero balance
	balance, err := request.CheckBalance(clientData.ID)
	if err != nil {
		NotifyClient(rw, "Unable to get balance", http.StatusInternalServerError)
		return
	}

	// Convert balance to another currency if needed
	if clientData.Currency != "" {
		clientData.ConvertedBalance, err = getConvertedBalance(clientData.Currency, balance)
		if err != nil {
			request.log.Println("[ERROR] Currency conversion failed. Reason: ", err)
			NotifyClient(rw, "Currency conversion failed", http.StatusInternalServerError)
			return
		}
	} else {
		clientData.Balance = balance
	}

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
	balance, err := request.CheckBalance(clientData.ID)
	if err != nil {
		NotifyClient(rw, "Unable to get balance for client", http.StatusInternalServerError)
		return
	}
	clientData.Balance = balance

	// Calculate new balance for current id
	balanceDif := clientData.Sum
	if balanceDif >= 0 {
		clientData.Balance += balanceDif
	} else {
		// make balance dif a positive number
		balanceDif *= -1

		// Do not allow to perform request if there is not enought balance
		if clientData.Balance >= balanceDif {
			clientData.Balance -= balanceDif
		} else {
			NotifyClient(rw, "Cannot satisfy request: not enough balance", http.StatusBadRequest)
			return
		}
	}

	// Update balance in database
	err = request.dbHandler.UpdateData(TableData{
		ClientID:  clientData.ID,
		Sum:       clientData.Balance,
		TableName: BALANCE_TABLE,
	})
	if err != nil {
		request.log.Printf("[ERROR] UPDATE query failed for ID %d. Reason: %s", clientData.ID, err)
		NotifyClient(rw, "Unable to update client data", http.StatusInternalServerError)
		return
	}

	if clientData.Description == "" {
		clientData.Description = "Update request"
	}
	request.LogCurrentTransaction(clientData.ID, clientData.Sum, clientData.Description)
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
	var sum int32
	if data.Sum < 0 {
		sum = -1 * data.Sum
	} else {
		sum = data.Sum
	}

	// If clinet's balance, who sends money to another client, is not enought to satisfy the request, send an error
	if senderBalance < sum {
		NotifyClient(rw, "Not enough balance to transfer money", http.StatusBadRequest)
		return
	}

	// Decrease balance of sender and increase balance of receiver
	err1 = request.dbHandler.UpdateData(TableData{
		ClientID:  data.ToID,
		Sum:       receiverBalance + sum,
		TableName: BALANCE_TABLE,
	})
	err2 = request.dbHandler.UpdateData(TableData{
		ClientID:  data.FromID,
		Sum:       senderBalance - sum,
		TableName: BALANCE_TABLE,
	})

	// If one of the balance updates returned an error, rollback successful update
	if err1 != nil || err2 != nil {
		if err1 == nil {
			request.dbHandler.UpdateData(TableData{
				ClientID:  data.ToID,
				Sum:       receiverBalance,
				TableName: BALANCE_TABLE,
			})
			request.log.Printf("[ERROR] UPDATE query failed for ID %d. Reason: %s", data.FromID, err2)
		} else if err2 == nil {
			request.dbHandler.UpdateData(TableData{
				ClientID:  data.FromID,
				Sum:       senderBalance,
				TableName: BALANCE_TABLE,
			})
			request.log.Printf("[ERROR] UPDATE query failed for ID %d. Reason: %s", data.ToID, err1)
		}

		NotifyClient(rw, "Unable to update client data", http.StatusInternalServerError)
	}

	if data.Description == "" {
		data.Description = fmt.Sprintf("Transfer money from %d to %d", data.FromID, data.ToID)
	}
	request.LogCurrentTransaction(data.ToID, sum, data.Description)
	request.LogCurrentTransaction(data.FromID, -sum, data.Description)
}

func (request *Request) GetTransactionsLog(rw http.ResponseWriter, r *http.Request) {

	clientData := r.Context().Value(KeyClient{}).(*ClientBalanceData)

	// Fetch logs from the database
	data, err := request.dbHandler.SelectData(TableData{
		ClientID:  clientData.ID,
		TableName: LOG_TABLE,
		Sort:      clientData.Sort,
	})
	if err != nil {
		request.log.Printf("[ERROR] SELECT query failed. Reason: %s", err)
		NotifyClient(rw, "Unable to get balance transactions logs", http.StatusInternalServerError)
	}

	// Send json logs to client
	for _, log := range data {
		err = log.ToJSON(rw)
		if err != nil {
			request.log.Println("[ERROR] Unable to parse JSON. Reason: ", err)
			NotifyClient(rw, "Unable to get balance", http.StatusInternalServerError)
			return
		}
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
