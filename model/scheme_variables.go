package model

import "encoding/json"

type SchemeVariable struct {
	Id      int32           `json:"id,omitempty" db:"id"`
	Name    string          `json:"name,omitempty" db:"name"`
	Encrypt bool            `json:"encrypt,omitempty" db:"encrypt"`
	Value   json.RawMessage `json:"value,omitempty" db:"value"`
}

type SearchSchemeVariable struct {
	ListRequest
}

var SchemeVariableFields = struct {
	Id      string
	Name    string
	Encrypt string
	Value   string
}{
	Id:      "id",
	Name:    "name",
	Encrypt: "encrypt",
	Value:   "value",
}

// TODO
func (v *SchemeVariable) IsValid() AppError {
	return nil
}
