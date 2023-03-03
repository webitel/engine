package sqlstore

import (
	"context"
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

func (s SqlUserStore) CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {

	res, err := s.GetReplica().WithContext(ctx).SelectNullInt(`select 1
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

func (s SqlUserStore) GetCallInfo(ctx context.Context, userId, domainId int64) (*model.UserCallInfo, *model.AppError) {
	var info *model.UserCallInfo
	err := s.GetReplica().WithContext(ctx).SelectOne(&info, `select u.id, coalesce( (u.name)::varchar, u.username) as name, 
u.extension, u.extension endpoint, d.name as domain_name, coalesce(u.profile, '{}'::jsonb) as variables, false as has_push
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

func (s SqlUserStore) GetCallInfoEndpoint(ctx context.Context, domainId int64, e *model.EndpointRequest, isOnline bool) (*model.UserCallInfo, *model.AppError) {
	var info *model.UserCallInfo
	err := s.GetReplica().WithContext(ctx).SelectOne(&info, `select u.id,
       coalesce((u.name)::varchar, u.username)                              as name,
       u.extension,
       u.extension                                                             endpoint,
       d.name                                                               as domain_name,
       coalesce(u.profile, '{}'::jsonb) || coalesce(push.config, '{}'::jsonb) as variables,
       exists(select 1
              from call_center.cc_agent a
                       left join call_center.cc_member_attempt aa on a.id = aa.agent_id
              where a.user_id = u.id::int8
                and (
                      (:IsOnline::bool is true and a.status != 'online')
                      or
                      (aa.leaving_at isnull
                          and now() - aa.last_state_change < interval '2m'
                          and aa.state in ('waiting_agent', 'idle', 'offering')
                          )
                  )
           )                                                                as is_busy,
		  push.config notnull as has_push
from directory.wbt_user u
         inner join directory.wbt_domain d on d.dc = u.dc
         left join lateral ( select jsonb_object(array_agg(key), array_agg(val)) as push
			from (SELECT case
							 when s.props ->> 'pn-type'::text = 'fcm' then 'wbt_push_fcm'
							 else 'wbt_push_apn' end                                            as key,
						 array_to_string(array_agg(DISTINCT s.props ->> 'pn-rpid'::text), '::') as val
				  FROM directory.wbt_session s
				  WHERE s.user_id IS NOT NULL
					AND s.access notnull
					AND NULLIF(s.props ->> 'pn-rpid'::text, ''::text) IS NOT NULL
					AND s.user_id = u.id
					and s.props ->> 'pn-type'::text in ('fcm', 'apns')
					AND now() at time zone 'UTC' < s.expires
				  group by s.props ->> 'pn-type'::text = 'fcm') t
			where key notnull
			  and val notnull) push(config) ON true
where case when :UserId::int8 notnull then u.id = :UserId else u.extension = :Extension::varchar end
  and u.dc = :DomainId
limit 1`, map[string]interface{}{
		"UserId":    e.UserId,
		"Extension": e.Extension,
		"DomainId":  domainId,
		"IsOnline":  isOnline,
	})

	if err != nil {
		return nil, model.NewAppError("SqlUserStore.GetCallInfoEndpoint", "store.sql_user.get_call_info.app_error", nil,
			fmt.Sprintf("UserId=%v, Extension=%v %s", e.UserId, e.Extension, err.Error()), extractCodeFromErr(err))
	}
	return info, nil
}

func (s SqlUserStore) DefaultWebRTCDeviceConfig(ctx context.Context, userId, domainId int64) (*model.UserDeviceConfig, *model.AppError) {
	var deviceConfig *model.UserDeviceConfig

	err := s.GetReplica().WithContext(ctx).SelectOne(&deviceConfig, `select u.extension,
       replace(coalesce(u.name, u.username), '"', '') display_name,
       dom.name as realm,
       'sip:' || u.extension || '@' || dom.name as uri,
       d.account as authorization_user,
       md5(d.account||':'||dom.name||':'||d.password) as ha1,
       '' as server
from directory.wbt_user u
    inner join directory.wbt_device d on d.id = u.device_id
    inner join directory.wbt_domain dom on dom.dc = u.dc
where u.id = :UserId and u.dc = :DomainId and u.extension notnull`, map[string]interface{}{
		"UserId":   userId,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlUserStore.DefaultDeviceConfig", "store.sql_user.get_default_device.app_error", nil,
			fmt.Sprintf("UserId=%v, %v", userId, err.Error()), extractCodeFromErr(err))
	}

	return deviceConfig, nil
}

func (s SqlUserStore) DefaultSipDeviceConfig(ctx context.Context, userId, domainId int64) (*model.UserSipDeviceConfig, *model.AppError) {
	var deviceConfig *model.UserSipDeviceConfig

	err := s.GetReplica().WithContext(ctx).SelectOne(&deviceConfig, `select u.extension,
       d.account as auth,
       dom.name as domain,
       coalesce(d.password, '') as password
from directory.wbt_user u
    inner join directory.wbt_device d on d.id = u.device_id
    inner join directory.wbt_domain dom on dom.dc = u.dc
where u.id = :UserId and u.dc = :DomainId and u.extension notnull`, map[string]interface{}{
		"UserId":   userId,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlUserStore.DefaultSipDeviceConfig", "store.sql_user.get_default_sip_device.app_error", nil,
			fmt.Sprintf("UserId=%v, %v", userId, err.Error()), extractCodeFromErr(err))
	}

	return deviceConfig, nil
}
