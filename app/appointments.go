package app

import (
	"context"
	"fmt"
	"time"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/utils"
	"golang.org/x/sync/singleflight"
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

func (app *App) GetAppointment(ctx context.Context, key string) (*model.Appointment, model.AppError) {
	if a, ok := cacheAppointments.Get(key); ok {
		return a.(*model.Appointment), nil
	}

	memberId, appErr := app.DecryptId(key)
	if appErr != nil {
		return nil, appErr
	}

	res, err, shared := appointmentGroupRequest.Do(fmt.Sprintf("member-%d", memberId), func() (interface{}, error) {
		a, err := app.Store.Member().GetAppointment(ctx, memberId)
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
		case model.AppError:
			return nil, err.(model.AppError)
		default:
			return nil, model.NewInternalError("app.appointment.get", err.Error())
		}
	}

	if !shared {
		cacheAppointments.AddWithDefaultExpires(key, res)
	}

	return res.(*model.Appointment), nil
}

func (app *App) AppointmentWidget(ctx context.Context, widgetUri string) (*model.AppointmentWidget, model.AppError) {
	if a, ok := cacheAppointmentDate.Get(widgetUri); ok {
		return a.(*model.AppointmentWidget), nil
	}

	return app.appointmentWidget(ctx, widgetUri)
}

func (app *App) appointmentWidget(ctx context.Context, widgetUri string) (*model.AppointmentWidget, model.AppError) {

	res, err, shared := appointmentGroupRequest.Do(fmt.Sprintf("list-%s", widgetUri), func() (interface{}, error) {
		a, err := app.Store.Member().GetAppointmentWidget(ctx, widgetUri)
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
		case model.AppError:
			return nil, err.(model.AppError)
		default:
			return nil, model.NewInternalError("app.appointment.list", err.Error())
		}
	}

	if !shared {
		cacheAppointmentDate.AddWithDefaultExpires(widgetUri, res)
	}

	return res.(*model.AppointmentWidget), nil
}

func (app *App) CreateAppointment(ctx context.Context, widget *model.AppointmentWidget, appointment *model.Appointment) (*model.Appointment, model.AppError) {
	var err model.AppError
	if !widget.ValidAppointment(appointment) {
		return nil, model.NewBadRequestError("appointment.valid.date", "No slot")
	}

	appointment, err = app.Store.Member().CreateAppointment(ctx, &widget.Profile, appointment)
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
	app.appointmentWidget(ctx, widget.Profile.Uri)

	return appointment, nil
}

func (app *App) CancelAppointment(ctx context.Context, widget *model.AppointmentWidget, key string) (*model.Appointment, model.AppError) {
	appointment, err := app.GetAppointment(ctx, key)
	if err != nil {
		return nil, err
	}

	if err = app.Store.Member().CancelAppointment(ctx, appointment.Id, "cancel"); err != nil {
		return nil, err
	}

	// reset list ?
	app.appointmentWidget(ctx, widget.Profile.Uri)

	return appointment, nil
}
