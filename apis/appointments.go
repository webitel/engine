package apis

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/webitel/engine/model"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	appointmentsCookie = "wbt_ac"
	errAllowOrigin     = model.NewAppError("API", "api.valid.origin", nil, "Not allow", http.StatusForbidden)
)

func (api *API) InitAppointments() {
	api.Routes.Root.Handle("/appointments/{id}", api.ApiHandlerTrustRequester(getAppointments)).Methods("GET")
	api.Routes.Root.Handle("/appointments/{id}", api.ApiHandlerTrustRequester(createAppointments)).Methods("POST")
	api.Routes.Root.Handle("/appointments/{id}", api.ApiHandlerTrustRequester(cancelAppointments)).Methods("DELETE")
}

/*
list/get

add
cancel
*/

func getAppointments(c *Context, w http.ResponseWriter, r *http.Request) {
	var cookie *http.Cookie
	var err error

	cookie, err = r.Cookie(appointmentsCookie)
	if err != nil && err != http.ErrNoCookie {
		w.WriteHeader(500)
		return
	}

	var appointment *model.Appointment

	widgetUri := getIdFromRequest(r)

	var widget *model.AppointmentWidget
	if widget, c.Err = c.App.AppointmentWidget(widgetUri); c.Err != nil {
		return
	}

	if !widget.AllowOrigin(c.IpAddress) {
		c.Err = errAllowOrigin
		return
	}

	if cookie != nil && cookie.Value != "" {
		if appointment, c.Err = c.App.GetAppointment(cookie.Value); c.Err != nil {
			if c.Err.StatusCode == http.StatusNotFound {
				cookie = &http.Cookie{
					Name:    appointmentsCookie,
					Value:   "",
					Path:    "/",
					Domain:  "",
					Expires: time.Unix(0, 0),
					MaxAge:  0,
				}
				http.SetCookie(w, cookie)
				c.Err = nil
			} else {
				return
			}

		} else {
			w.Write(appointment.Computed)
			return
		}
	}

	w.Write(widget.ComputedList)

	return
}

func createAppointments(c *Context, w http.ResponseWriter, r *http.Request) {
	var cookie *http.Cookie
	var body []byte
	var err error

	cookie, err = r.Cookie(appointmentsCookie)
	if err != nil && err != http.ErrNoCookie {
		w.WriteHeader(500)
		return
	}

	if cookie != nil && cookie.Value != "" {
		getAppointments(c, w, r)
		return
	}

	body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		//todo error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	widgetUri := getIdFromRequest(r)

	var widget *model.AppointmentWidget
	if widget, c.Err = c.App.AppointmentWidget(widgetUri); c.Err != nil {
		return
	}

	if !widget.AllowOrigin(c.IpAddress) {
		c.Err = errAllowOrigin
		return
	}

	appointment := model.AppointmentFromJson(body)
	appointment.Ip = c.IpAddress

	if appointment.Variables == nil {
		appointment.Variables = model.StringMap{}
	}

	if ua := r.Header.Get("User-Agent"); ua != "" {
		appointment.Variables["user_agent"] = ua
	}
	appointment.Variables["origin"] = r.Header.Get("Origin")
	appointment, c.Err = c.App.CreateAppointment(widget, appointment)
	if c.Err != nil {
		return
	}

	expires, _ := time.Parse("2006-01-02T15:04:05", fmt.Sprintf("%sT23:59:59", appointment.ScheduleDate))

	cookie = &http.Cookie{
		Name:    appointmentsCookie,
		Value:   appointment.Key,
		Path:    "/",
		Domain:  "",
		Expires: expires,
		MaxAge:  0, // TODO CONFIG
	}
	http.SetCookie(w, cookie)
	w.Write(appointment.Computed)
}

func cancelAppointments(c *Context, w http.ResponseWriter, r *http.Request) {
	var cookie *http.Cookie
	var err error

	cookie, err = r.Cookie(appointmentsCookie)
	if err != nil && err != http.ErrNoCookie {
		w.WriteHeader(500)
		return
	}

	// TODO
	if cookie == nil {
		w.WriteHeader(500)
		return
	}

	widgetUri := getIdFromRequest(r)

	var widget *model.AppointmentWidget
	if widget, c.Err = c.App.AppointmentWidget(widgetUri); c.Err != nil {
		return
	}

	if !widget.AllowOrigin(c.IpAddress) {
		c.Err = errAllowOrigin
		return
	}

	if _, c.Err = c.App.CancelAppointment(widget, cookie.Value); c.Err != nil {
		return
	}

	cookie = &http.Cookie{
		Name:    appointmentsCookie,
		Value:   "",
		Path:    "/",
		Domain:  "",
		Expires: time.Unix(0, 0),
		MaxAge:  0,
	}
	http.SetCookie(w, cookie)
	w.Write(widget.ComputedList)
}

func getIdFromRequest(r *http.Request) string {
	props := mux.Vars(r)
	return "/" + props["id"]
}
