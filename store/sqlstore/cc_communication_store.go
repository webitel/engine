package sqlstore

import (
	"context"
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlCommunicationTypeStore struct {
	SqlStore
}

func NewSqlCommunicationTypeStore(sqlStore SqlStore) store.CommunicationTypeStore {
	us := &SqlCommunicationTypeStore{sqlStore}
	return us
}

func (s SqlCommunicationTypeStore) Create(ctx context.Context, comm *model.CommunicationType) (*model.CommunicationType, *model.AppError) {
	var out *model.CommunicationType
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `insert into call_center.cc_communication (name, code, type, domain_id, description)
		values (:Name, :Code, :Type, :DomainId, :Description)
		returning *`,
		map[string]interface{}{
			"Name":        comm.Name,
			"Code":        comm.Code,
			"Type":        comm.Type,
			"DomainId":    comm.DomainId,
			"Description": comm.Description,
		}); nil != err {
		return nil, model.NewAppError("SqlCommunicationTypeStore.Save", "store.sql_communication_type.save.app_error", nil,
			fmt.Sprintf("name=%v, %v", comm.Name, err.Error()), http.StatusInternalServerError)
	} else {
		return out, nil
	}
}

func (s SqlCommunicationTypeStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchCommunicationType) ([]*model.CommunicationType, *model.AppError) {
	var communications []*model.CommunicationType

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(ctx, &communications, search.ListRequest,
		`domain_id = :DomainId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar))`,
		model.CommunicationType{}, f)

	if err != nil {
		return nil, model.NewAppError("SqlCommunicationTypeStore.GetAllPage", "store.sql_communication_type.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return communications, nil
	}
}

func (s SqlCommunicationTypeStore) Get(ctx context.Context, domainId int64, id int64) (*model.CommunicationType, *model.AppError) {
	var out *model.CommunicationType
	if err := s.GetReplica().WithContext(ctx).SelectOne(&out, `select *
		from call_center.cc_communication s
		where s.id = :Id and s.domain_id = :DomainId`, map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return nil, model.NewAppError("SqlCommunicationTypeStore.Get", "store.sql_communication_type.get.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlCommunicationTypeStore) Update(ctx context.Context, cType *model.CommunicationType) (*model.CommunicationType, *model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&cType, `update call_center.cc_communication
set name = :Name,
    description = :Description,
    type = :Type,
    code = :Code
where id = :Id and domain_id = :DomainId
returning *`, map[string]interface{}{
		"Name":        cType.Name,
		"Description": cType.Description,
		"Type":        cType.Type,
		"Code":        cType.Code,
		"Id":          cType.Id,
		"DomainId":    cType.DomainId,
	})
	if err != nil {
		return nil, model.NewAppError("SqlCommunicationTypeStore.Update", "store.sql_communication_type.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", cType.Id, err.Error()), http.StatusInternalServerError)
	}
	return cType, nil
}

func (s SqlCommunicationTypeStore) Delete(ctx context.Context, domainId int64, id int64) *model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_communication c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlCommunicationTypeStore.Delete", "store.sql_communication_type.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}
