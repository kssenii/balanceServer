package handlers

import (
	"encoding/json"
	"io"

	"gopkg.in/go-playground/validator.v9"
)

type ClientBalanceData struct {
	ID      uint64 `json:"id" validate:"required"`
	Balance uint64 `json:"balance" validate:"gte=0"`
	Sum     uint64 `json:"sum" validate:"gte=0"`
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
	err := validate.Struct(data)
	return err
}
