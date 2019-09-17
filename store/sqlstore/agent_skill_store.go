package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlSkillStore struct {
	SqlStore
}

func NewSqlSkillStore(sqlStore SqlStore) store.SkillStore {
	us := &SqlSkillStore{sqlStore}
	return us
}

func (s SqlSkillStore) Create(skill *model.Skill) (*model.Skill, *model.AppError) {
	var out *model.Skill
	if err := s.GetMaster().SelectOne(&out, `insert into cc_skill (name, domain_id, description)
		values (:Name, :DomainId, :Description)
		returning *`,
		map[string]interface{}{"Name": skill.Name, "DomainId": skill.DomainId, "Description": skill.Description}); nil != err {
		return nil, model.NewAppError("SqlSkillStore.Save", "store.sql_skill.save.app_error", nil,
			fmt.Sprintf("name=%v, %v", skill.Name, err.Error()), http.StatusInternalServerError)
	} else {
		return out, nil
	}
}

func (s SqlSkillStore) Get(domainId int64, id int64) (*model.Skill, *model.AppError) {
	var skill *model.Skill
	if err := s.GetReplica().SelectOne(&skill, `select *
		from cc_skill s
		where s.id = :Id and s.domain_id = :DomainId`, map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("SqlSkillStore.Get", "store.sql_skill.get.app_error", nil,
				fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusNotFound)
		} else {
			return nil, model.NewAppError("SqlCalendarStore.Get", "store.sql_skill.get.app_error", nil,
				fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
		}
	} else {
		return skill, nil
	}
}

func (s SqlSkillStore) GetAllPage(domainId int64, offset, limit int) ([]*model.Skill, *model.AppError) {
	var skills []*model.Skill

	if _, err := s.GetReplica().Select(&skills,
		`select c.id,
       c.name,
       c.description
from cc_skill c
where c.domain_id = :DomainId
order by id
limit :Limit
offset :Offset`, map[string]interface{}{"DomainId": domainId, "Limit": limit, "Offset": offset}); err != nil {
		return nil, model.NewAppError("SqlSkillStore.GetAllPage", "store.sql_skill.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return skills, nil
	}
}

func (s SqlSkillStore) Delete(domainId int64, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from cc_skill c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlSkillStore.Delete", "store.sql_skill.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}

func (s SqlSkillStore) Update(skill *model.Skill) (*model.Skill, *model.AppError) {
	err := s.GetMaster().SelectOne(&skill, `update cc_skill
	set name = :Name,
    description = :Description
		where id = :Id and domain_id = :DomainId returning *`, map[string]interface{}{
		"Id":          skill.Id,
		"Name":        skill.Name,
		"Description": skill.Description,
		"DomainId":    skill.DomainId,
	})
	if err != nil {
		return nil, model.NewAppError("SqlSkillStore.Update", "store.sql_skill.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", skill.Id, err.Error()), http.StatusInternalServerError)
	}
	return skill, nil
}
