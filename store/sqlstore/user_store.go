package sqlstore

import (
	"fmt"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlUserStore struct {
	SqlStore
}

func NewSqlUserStore(sqlStore SqlStore) store.UserStore {
	us := &SqlUserStore{sqlStore}
	return us
}

func (s SqlUserStore) GetCallInfo(userId, domainId int64) (*model.UserCallInfo, *model.AppError) {
	var info *model.UserCallInfo
	err := s.GetReplica().SelectOne(&info, `select u.name, u.extension, d.name as domain_name, u.envars as variables
from directory.wbt_user u
    inner join directory.wbt_domain d on d.dc = u.dc
where u.id = :UserId
  and u.dc = :DomainId`, map[string]interface{}{
		"UserId":   userId,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlUserStore.GetCallInfo", "store.sql_user.get_call_info.app_error", nil,
			fmt.Sprintf("UserId=%v, %s", userId, err.Error()), extractCodeFromErr(err))
	}
	return info, nil
}
