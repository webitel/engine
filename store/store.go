package store

import (
	"time"

	"github.com/webitel/engine/model"
)

type StoreResult struct {
	Data interface{}
	Err  *model.AppError
}

type StoreChannel chan StoreResult

func Do(f func(result *StoreResult)) StoreChannel {
	storeChannel := make(StoreChannel, 1)
	go func() {
		result := StoreResult{}
		f(&result)
		storeChannel <- result
		close(storeChannel)
	}()
	return storeChannel
}

func Must(sc StoreChannel) interface{} {
	r := <-sc
	if r.Err != nil {

		time.Sleep(time.Second)
		panic(r.Err)
	}

	return r.Data
}

type Store interface {
	Session() SessionStore
	Calendar() CalendarStore
}

type SessionStore interface {
	Get(sessionIdOrToken string) StoreChannel
}

type CalendarStore interface {
	Create(calendar *model.Calendar) (*model.Calendar, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.Calendar, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, offset, limit int) ([]*model.Calendar, *model.AppError)
	Get(domainId int64, id int64) (*model.Calendar, *model.AppError)
	Update(calendar *model.Calendar) (*model.Calendar, *model.AppError)
	Delete(domainId, id int64) *model.AppError

	GetTimezoneAllPage(offset, limit int) ([]*model.Timezone, *model.AppError)

	CheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError)

	GetAcceptOfDay(calendarId int64) ([]*model.CalendarAcceptOfDay, *model.AppError)
}
