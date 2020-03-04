package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

type User struct {
	Id       *int64  `json:"id"`
	DomainId *int64  `json:"domain_id"`
	Token    string  `json:"token"` //todo wbt_token -> access
	Name     string  `json:"name"`
	GroupIds []int64 `json:"group_ids"`
}

type UserCallInfo struct {
	Id         int64              `json:"id" db:"id"`
	Name       string             `json:"name" db:"name"`
	DomainName string             `json:"domain_name" db:"domain_name"`
	Extension  string             `json:"tel_number" db:"extension"`
	Endpoint   string             `json:"endpoint" db:"endpoint"`
	Variables  *map[string]string `json:"variables" db:"variables"`
}

func (u *UserCallInfo) GetCallEndpoints() []string {
	return []string{fmt.Sprintf("sofia/sip/%s@%s", u.Endpoint, u.DomainName)}
}

func (u UserCallInfo) GetVariables() map[string]string {
	if u.Variables != nil {
		return *u.Variables
	}

	return make(map[string]string)
}

func (u UserCallInfo) BridgeEndpoint() string {
	return strings.Join(u.GetCallEndpoints(), ",")
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

func UnionStringMaps(src ...map[string]string) map[string]string {
	res := make(map[string]string)
	for _, m := range src {
		for k, v := range m {
			res[k] = v
		}
	}
	return res
}
