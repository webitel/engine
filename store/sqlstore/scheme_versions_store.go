package sqlstore

import (
	"context"
	"fmt"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlSchemeVersionsStore struct {
	SqlStore
}

func NewSqlSchemeVersionsStore(sqlStore SqlStore) store.SchemeVersionsStore {
	us := &SqlSchemeVersionsStore{sqlStore}
	return us
}

func (s SqlSchemeVersionsStore) Get(ctx context.Context, flowId int32) ([]*model.SchemaVersion, model.AppError) {
	var versions []*model.SchemaVersion
	err := s.GetReplica().WithContext(ctx).SelectOne(&versions, `
	select s.id, s.scheme_id, s.created_at,  call_center.cc_get_lookup(s.created_by::bigint,usr.name::varchar) created_by, s.scheme, s.payload, s.version, s.note
		from call_center.scheme_version s
		left join directory.wbt_user usr on usr.id = s.created_by
	where s.scheme_id = :Id
`, map[string]interface{}{
		"Id": flowId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_scheme_version.get.app_error", fmt.Sprintf("Flow Id=%v, %s", flowId, err.Error()), extractCodeFromErr(err))
	}

	return versions, nil
}
