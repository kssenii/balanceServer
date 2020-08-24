package handlers

import (
	"encoding/json"
	"io"

	"gopkg.in/go-playground/validator.v9"
)

type ClientBalanceData struct {
	ID      uint32 `json:"id"`            // client id
	Sum     int32  `json:"sum,omitempty"` // sum to add to current balance
	Balance int32  `json:"balance"`       // current balance

	ToID   uint32 `json:"toID,omitempty" validate:"required_with=FromID"` // client's id, to whom sum is transfered
	FromID uint32 `json:"fromID,omitempty" validate:"required_with=ToID"` // client's id, from whom sum is taken

	Currency         string  `json:"currency,omitempty"`
	ConvertedBalance float64 `json:"converted_balance,omitempty"`
	Description      string  `json:"description,omitempty"`
	Sort             string  `json:"sort,omitempty"`
}

type TableData struct {
	ClientID    uint32 `json:"id"`
	Time        string `json:"time"`
	Sum         int32  `json:"sum"`
	Description string `json:"description"`
	TableName   string `json:"-"`
	Sort        string `json:"-"`
}

func (data *ClientBalanceData) ToJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(data)
}

func (data *ClientBalanceData) FromJSON(r io.Reader) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(data)
}

func (data *ClientBalanceData) Validate() error {
	validate := validator.New()
	return validate.Struct(data)
}

func (data *TableData) ToJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(data)
}
