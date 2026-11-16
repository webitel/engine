package apis

import (
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/webitel/engine/model"
)

const appointmentHeader = "X-WBT-KEY"

var (
	errAllowOrigin = model.NewForbiddenError("api.valid.origin", "Not allow")
)

func (api *API) InitAppointments() {
	api.Routes.Root.Handle("/appointments/{id}", api.ApiHandlerTrustRequester(listAppointments)).Methods("GET")
	api.Routes.Root.Handle("/appointments/{id}", api.ApiHandlerTrustRequester(createAppointments)).Methods("POST")
	api.Routes.Root.Handle("/appointments/{id}", api.ApiHandlerTrustRequester(cancelAppointments)).Methods("DELETE")
}

func listAppointments(c *Context, w http.ResponseWriter, r *http.Request) {
	uri := getIdFromRequest(r)

	widget, err := c.App.AppointmentWidget(r.Context(), uri)
	if err != nil {
		c.Err = err

		return
	}

	if !widget.AllowOrigin(c.IpAddress) {
		c.Err = errAllowOrigin

		return
	}

	key, _ := getKeyFromRequest(r)
	if strings.TrimSpace(key) == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Write(widget.ComputedList)

		return
	}

	appointment, err := c.App.GetAppointment(r.Context(), key, widget.Profile.DomainId, int64(widget.Profile.QueueId))
	if err != nil {
		if err.GetStatusCode() == http.StatusNotFound {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Write(widget.ComputedList)

			return
		}

		c.Err = err
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(appointment.Computed)
}

func createAppointments(c *Context, w http.ResponseWriter, r *http.Request) {
	var body []byte
	var err error

	key, _ := getKeyFromRequest(r)

	if key != "" {
		listAppointments(c, w, r)
		return
	}

	body, err = io.ReadAll(r.Body)
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
	return "/" + mux.Vars(r)["id"]
}

func getKeyFromRequest(r *http.Request) (string, bool) {
	header := r.Header.Get(appointmentHeader)
	if header != "" {
		return header, true
	}

	return "", false
}
