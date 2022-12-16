package app

import (
	"fmt"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/utils"
	"golang.org/x/sync/singleflight"
	"net/http"
	"time"
)

const (
	sizeCacheAppointments = 10000
)

var (
	cacheAppointments       utils.ObjectCache
	cacheAppointmentDate    utils.ObjectCache
	appointmentGroupRequest singleflight.Group
)

func init() {
	cacheAppointments = utils.NewLruWithParams(sizeCacheAppointments, "Appointment", 60, "")
	cacheAppointmentDate = utils.NewLruWithParams(sizeCacheAppointments, "List appointment date", 60, "")
}

func (app *App) GetAppointment(key string) (*model.Appointment, *model.AppError) {
	if a, ok := cacheAppointments.Get(key); ok {
		return a.(*model.Appointment), nil
	}

	memberId, appErr := app.DecryptId(key)
	if appErr != nil {
		return nil, appErr
	}

	res, err, shared := appointmentGroupRequest.Do(fmt.Sprintf("member-%d", memberId), func() (interface{}, error) {
		a, err := app.Store.Member().GetAppointment(memberId)
		if err != nil {
			return nil, err
		}

		a.Computed = (&model.AppointmentResponse{
			Timezone:    a.Timezone,
			Type:        "appointment",
			List:        nil,
			Appointment: a,
		}).ToJSON()

		return a, nil
	})

	if err != nil {
		switch err.(type) {
		case *model.AppError:
			return nil, err.(*model.AppError)
		default:
			return nil, model.NewAppError("App", "app.appointment.get", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	if !shared {
		cacheAppointments.AddWithDefaultExpires(key, res)
	}

	return res.(*model.Appointment), nil
}

func (app *App) AppointmentWidget(widgetUri string) (*model.AppointmentWidget, *model.AppError) {
	if a, ok := cacheAppointmentDate.Get(widgetUri); ok {
		return a.(*model.AppointmentWidget), nil
	}

	return app.appointmentWidget(widgetUri)
}

func (app *App) appointmentWidget(widgetUri string) (*model.AppointmentWidget, *model.AppError) {

	res, err, shared := appointmentGroupRequest.Do(fmt.Sprintf("list-%s", widgetUri), func() (interface{}, error) {
		a, err := app.Store.Member().GetAppointmentWidget(widgetUri)
		if err != nil {
			return nil, err
		}
		a.Loc, _ = time.LoadLocation(a.Profile.Timezone)
		a.ComputedList = (&model.AppointmentResponse{
			Timezone: a.Profile.Timezone,
			Type:     "list",
			List:     a.List,
		}).ToJSON()

		return a, nil
	})

	if err != nil {
		switch err.(type) {
		case *model.AppError:
			return nil, err.(*model.AppError)
		default:
			return nil, model.NewAppError("App", "app.appointment.list", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	if !shared {
		cacheAppointmentDate.AddWithDefaultExpires(widgetUri, res)
	}

	return res.(*model.AppointmentWidget), nil
}

func (app *App) CreateAppointment(widget *model.AppointmentWidget, appointment *model.Appointment) (*model.Appointment, *model.AppError) {
	var err *model.AppError
	if !widget.ValidAppointment(appointment) {
		return nil, model.NewAppError("CreateAppointment", "appointment.valid.date", nil, "No slot", http.StatusBadRequest)
	}

	appointment, err = app.Store.Member().CreateAppointment(&widget.Profile, appointment)
	if err != nil {
		return nil, err
	}

	appointment.Key, err = app.EncryptId(appointment.Id)
	if err != nil {
		return nil, err
	}

	appointment.Timezone = widget.Profile.Timezone

	expires, _ := time.Parse("2006-01-02T15:04:05", fmt.Sprintf("%sT23:59:59", appointment.ScheduleDate))
	if widget.Loc != nil {
		expires.In(widget.Loc)
	}
	appointment.ExpireKey = expires.UnixMilli()

	appointment.Computed = (&model.AppointmentResponse{
		Timezone:    widget.Profile.Timezone,
		Type:        "appointment",
		List:        nil,
		Appointment: appointment,
	}).ToJSON()

	// reset list ?
	app.appointmentWidget(widget.Profile.Uri)

	return appointment, nil
}

func (app *App) CancelAppointment(widget *model.AppointmentWidget, key string) (*model.Appointment, *model.AppError) {
	appointment, err := app.GetAppointment(key)
	if err != nil {
		return nil, err
	}

	if err = app.Store.Member().CancelAppointment(appointment.Id, "cancel"); err != nil {
		return nil, err
	}

	// reset list ?
	app.appointmentWidget(widget.Profile.Uri)

	return appointment, nil
}
