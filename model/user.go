package model

import (
	"encoding/json"
	"fmt"
)

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
	Extension  string                  `json:"tel_number" db:"extension"`
	Variables  *map[string]interface{} `json:"variables" db:"variables"`
}

func (u *UserCallInfo) GetCallEndpoints() []string {
	return []string{fmt.Sprintf("sofia/sip/%s@%s", u.Extension, u.DomainName)}
}

func (u *User) Root() bool {
	return u.Id == nil && u.DomainId == nil
}

type UserDeviceConfig struct {
	Server            string `json:"server" db:"server"`
	Extension         string `json:"extension" db:"extension"`
	Realm             string `json:"realm" db:"realm"`
	Uri               string `json:"uri" db:"uri"`
	AuthorizationUser string `json:"authorization_user" db:"authorization_user"`
	Ha1               string `json:"ha1" db:"ha1"`
}

func (d UserDeviceConfig) ToMap() map[string]interface{} {
	out := make(map[string]interface{})
	data, _ := json.Marshal(d)
	_ = json.Unmarshal(data, &out)
	return out
}
