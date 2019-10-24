package model

import "fmt"

type User struct {
	Id       *int64  `json:"id"`
	DomainId *int64  `json:"domain_id"`
	Token    string  `json:"token"` //todo wbt_token -> access
	Name     string  `json:"name"`
	GroupIds []int64 `json:"group_ids"`
}

type UserCallInfo struct {
	Name       string                  `json:"name" db:"name"`
	DomainName string                  `json:"domain_name" db:"domain_name"`
	TelNumber  string                  `json:"tel_number" db:"tel_number"`
	Variables  *map[string]interface{} `json:"variables" db:"variables"`
}

func (u *UserCallInfo) GetCallEndpoints() []string {
	return []string{fmt.Sprintf("sofia/sip/%s@%s", u.TelNumber, u.DomainName)}
}

func (u *User) Root() bool {
	return u.Id == nil && u.DomainId == nil
}
