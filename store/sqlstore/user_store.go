package sqlstore

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/auth_manager"
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

func (s SqlUserStore) CheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {

	res, err := s.GetReplica().SelectNullInt(`select 1
		where exists(
          select 1
          from directory.wbt_auth_acl a
          where a.dc = :DomainId
            and a.object = :Id
            and a.subject = any (:Groups::int[])
            and a.access & :Access = :Access
        )`, map[string]interface{}{"DomainId": domainId, "Id": id, "Groups": pq.Array(groups), "Access": access.Value()})

	if err != nil {
		return false, nil
	}

	return res.Valid && res.Int64 == 1, nil
}

func (s SqlUserStore) GetCallInfo(userId, domainId int64) (*model.UserCallInfo, *model.AppError) {
	var info *model.UserCallInfo
	err := s.GetReplica().SelectOne(&info, `select coalesce( (u.name)::varchar, u.username) as name, u.extension, u.username endpoint, d.name as domain_name, u.profile as variables
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

func (s SqlUserStore) DefaultDeviceConfig(userId, domainId int64) (*model.UserDeviceConfig, *model.AppError) {
	var deviceConfig *model.UserDeviceConfig

	err := s.GetReplica().SelectOne(&deviceConfig, `select u.extension,
       dom.name as realm,
       'sip:' || u.username || '@' || dom.name as uri,
       d.account as authorization_user,
       d.password as ha1,
       '' as server
from directory.wbt_user u
    inner join directory.wbt_device d on d.id = u.device_id
    inner join directory.wbt_domain dom on dom.dc = u.dc
where u.id = :UserId and u.dc = :DomainId`, map[string]interface{}{
		"UserId":   userId,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlUserStore.DefaultDeviceConfig", "store.sql_user.get_default_device.app_error", nil,
			fmt.Sprintf("UserId=%v, %v", userId, err.Error()), extractCodeFromErr(err))
	}

	return deviceConfig, nil
}
