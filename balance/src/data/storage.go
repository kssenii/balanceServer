package data

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type Client struct {
	ID        uint64 `json:"id"`
	Balance   uint64 `json:"balance"`
	CreatedOn string `json:"-"`
	UpdatedOn string `json:"-"`
}

type Clients []*Client

func (clients *Clients) ToJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(clients)
}

func (clients *Client) FromJSON(r io.Reader) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(clients)
}

func GetCLients() Clients {
	return clientsList
}

func AddClient(client *Client) {
	client.ID = GetNextID()
	clientsList = append(clientsList, client)
}

func GetNextID() uint64 {
	clientList := clientsList[len(clientsList)-1]
	return clientList.ID + 1
}

func UpdateClient(id uint64, client *Client) error {
	pos, err := findClient(id)

	if err != nil {
		return err
	}

	client.ID = id
	clientsList[pos] = client

	return nil
}

var ErrClientNotFound = fmt.Errorf("Client ID not found")

func findClient(id uint64) (int, error) {
	for i, cl := range clientsList {
		if cl.ID == id {
			return i, nil
		}
	}

	return -1, ErrClientNotFound
}

var clientsList = []*Client{
	&Client{
		ID:        1,
		Balance:   0,
		CreatedOn: time.Now().UTC().String(),
		UpdatedOn: time.Now().UTC().String(),
	},
	&Client{
		ID:        2,
		Balance:   0,
		CreatedOn: time.Now().UTC().String(),
		UpdatedOn: time.Now().UTC().String(),
	},
}
