package apis

import (
	"github.com/gorilla/mux"
	"github.com/webitel/engine/model"
	"io/ioutil"
	"net/http"
)

var (
	appointmentHeader = "X-WBT-KEY"

	errAllowOrigin = model.NewAppError("API", "api.valid.origin", nil, "Not allow", http.StatusForbidden)
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

	key, _ := getKeyFromRequest(r)

	var appointment *model.Appointment

	widgetUri := getIdFromRequest(r)
	ctx := r.Context()

	var widget *model.AppointmentWidget
	if widget, c.Err = c.App.AppointmentWidget(ctx, widgetUri); c.Err != nil {
		return
	}

	if !widget.AllowOrigin(c.IpAddress) {
		c.Err = errAllowOrigin
		return
	}

	if key != "" {
		if appointment, c.Err = c.App.GetAppointment(ctx, key); c.Err != nil {
			if c.Err.StatusCode == http.StatusNotFound {
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
	var body []byte
	var err error

	key, _ := getKeyFromRequest(r)

	if key != "" {
		getAppointments(c, w, r)
		return
	}

	body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		//todo error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	widgetUri := getIdFromRequest(r)

	var widget *model.AppointmentWidget
	if widget, c.Err = c.App.AppointmentWidget(r.Context(), widgetUri); c.Err != nil {
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
	appointment, c.Err = c.App.CreateAppointment(r.Context(), widget, appointment)
	if c.Err != nil {
		return
	}

	w.Header().Set(appointmentHeader, appointment.Key)
	w.Write(appointment.Computed)
}

func cancelAppointments(c *Context, w http.ResponseWriter, r *http.Request) {
	key, _ := getKeyFromRequest(r)
	if key == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	widgetUri := getIdFromRequest(r)

	var widget *model.AppointmentWidget
	if widget, c.Err = c.App.AppointmentWidget(r.Context(), widgetUri); c.Err != nil {
		return
	}

	if !widget.AllowOrigin(c.IpAddress) {
		c.Err = errAllowOrigin
		return
	}

	if _, c.Err = c.App.CancelAppointment(r.Context(), widget, key); c.Err != nil {
		return
	}
	w.Write(widget.ComputedList)
}

func getIdFromRequest(r *http.Request) string {
	props := mux.Vars(r)
	return "/" + props["id"]
}

func getKeyFromRequest(r *http.Request) (string, bool) {
	header := r.Header.Get(appointmentHeader)
	if header != "" {
		return header, true
	}

	return "", false
}
