package data

import (
	"encoding/json"
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

func GetCLients() Clients {
	return clientsList
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
