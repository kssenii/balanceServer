package handlers

import (
	"encoding/json"
	"io"

	"gopkg.in/go-playground/validator.v9"
)

type ClientBalanceData struct {
	ID      uint32 `json:"id"`                                             // client id
	Balance uint64 `json:"Balance"`                                        // current balance
	Sum     int32  `json:"sum,omitempty"`                                  // sum to add to current balance
	ToID    uint32 `json:"toID,omitempty" validate:"required_with=FromID"` // client's id, to whom sum is transfered
	FromID  uint32 `json:"fromID,omitempty" validate:"required_with=ToID"` // client's id, from whom sum is taken
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
