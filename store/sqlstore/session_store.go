package sqlstore

import (
	"fmt"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlSessionStore struct {
	SqlStore
}

func NewSqlSessionStore(sqlStore SqlStore) store.SessionStore {
	us := &SqlSessionStore{sqlStore}
	return us
}

func (sql SqlSessionStore) Get(token string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var session *model.Session
		err := sql.GetReplica().SelectOne(&session, `select  t.id,
       wa.id as user_id,
       wa.dc                                                                      as domain_id,
       wa.caller_name                                                             as name,
       extract(epoch from coalesce(t.rotated, t.created)::timestamp + (t.expires * '1 second'::interval))::bigint as expires_at,
	   t.access as token
from wbt_token t
       inner join wbt_auth wa on t.domain_id = wa.dc and t.owner_id = wa.id
where t.access = :Token`, map[string]interface{}{"Token": token})
		if err != nil {
			result.Err = model.NewAppError("SqlSessionStore.Get", "store.sql_session.get.app_error", nil,
				fmt.Sprintf("token=%v, %s", token, err.Error()), http.StatusNotFound)
		} else {
			result.Data = session
		}
	})
}
