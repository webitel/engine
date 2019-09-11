package model

type User struct {
	Id       *int64  `json:"id"`
	DomainId *int64  `json:"domain_id"`
	Token    string  `json:"token"` //todo wbt_token -> access
	Name     string  `json:"name"`
	GroupIds []int64 `json:"group_ids"`
}

func (u *User) Root() bool {
	return u.Id == nil && u.DomainId == nil
}
