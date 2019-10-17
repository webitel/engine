package sqlstore

import (
	"fmt"
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

func (s SqlCommunicationTypeStore) Create(comm *model.CommunicationType) (*model.CommunicationType, *model.AppError) {
	var out *model.CommunicationType
	if err := s.GetMaster().SelectOne(&out, `insert into cc_communication (name, code, type, domain_id, description)
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

func (s SqlCommunicationTypeStore) GetAllPage(domainId int64, offset, limit int) ([]*model.CommunicationType, *model.AppError) {
	var communications []*model.CommunicationType

	if _, err := s.GetReplica().Select(&communications,
		`select id, name, code, description, type
from cc_communication c
where c.domain_id = :DomainId
order by id
limit :Limit
offset :Offset`, map[string]interface{}{"DomainId": domainId, "Limit": limit, "Offset": offset}); err != nil {
		return nil, model.NewAppError("SqlCommunicationTypeStore.GetAllPage", "store.sql_communication_type.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return communications, nil
	}
}

func (s SqlCommunicationTypeStore) Get(domainId int64, id int64) (*model.CommunicationType, *model.AppError) {
	var out *model.CommunicationType
	if err := s.GetReplica().SelectOne(&out, `select *
		from cc_communication s
		where s.id = :Id and s.domain_id = :DomainId`, map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return nil, model.NewAppError("SqlCommunicationTypeStore.Get", "store.sql_communication_type.get.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlCommunicationTypeStore) Update(cType *model.CommunicationType) (*model.CommunicationType, *model.AppError) {
	err := s.GetMaster().SelectOne(&cType, `update cc_communication
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

func (s SqlCommunicationTypeStore) Delete(domainId int64, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from cc_communication c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlCommunicationTypeStore.Delete", "store.sql_communication_type.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}
