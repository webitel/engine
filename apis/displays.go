package apis

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/webitel/engine/pkg/wbt/auth_manager"

	"github.com/gorilla/mux"
	"github.com/webitel/engine/model"
)

func (api *API) InitDisplays() {
	api.Routes.Root.Handle("/api/displays/{id}", api.ApiHandlerRequester(createDisplays)).Methods("POST")
}

// see https://stackoverflow.com/a/21375405
var bom = []byte{0xef, 0xbb, 0xbf}

func trimBOM(data []byte) []byte {
	if bytes.HasPrefix(data, bom) {
		return data[len(bom):]
	}
	return data
}

func createDisplays(c *Context, w http.ResponseWriter, r *http.Request) {
	session := &c.Session
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE)
	if !permission.CanRead() {
		c.Err = c.App.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
		return
	}

	if !permission.CanUpdate() {
		c.Err = c.App.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		return
	}

	resourceId, err := parseResourceID(r)
	if err != nil {
		c.Err = model.NewBadRequestError("api.displays.valid.resourceId", "Invalid resource ID: "+err.Error())
		return
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		if perm, appErr := c.App.OutboundResourceCheckAccess(r.Context(), session.Domain(0), resourceId, session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); appErr != nil {
			c.Err = appErr
			return
		} else if !perm {
			c.Err = c.App.MakeResourcePermissionError(session, resourceId, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
			return
		}
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
		c.Err = model.NewBadRequestError("api.displays.valid.file", "Error parsing form data: "+err.Error())
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		c.Err = model.NewBadRequestError("api.displays.valid.file", "Error retrieving file: "+err.Error())
		return
	}
	defer file.Close()

	delimiter := getDelimiter(r.FormValue("delimiter"))
	mapCol := r.FormValue("map")

	records, err := readCSV(file, delimiter)
	if err != nil {
		c.Err = model.NewBadRequestError("api.displays.valid.file", "Error reading CSV file: "+err.Error())
		return
	}

	if len(records) == 0 {
		c.Err = model.NewBadRequestError("api.displays.valid.file", "CSV file is empty")
		return
	}

	columnIndex, err := findColumnIndex(records[0], mapCol)
	if err != nil {
		c.Err = model.NewBadRequestError("api.displays.valid.file", err.Error())
		return
	}

	mappedData, werr := mapData(records, columnIndex, resourceId)
	if c.Err = werr; c.Err != nil {
		return
	}

	displays, appErr := c.App.CreateOutboundResourceDisplays(r.Context(), resourceId, mappedData)
	if appErr != nil {
		c.Err = appErr
		return
	}

	writeJSONResponse(w, displays)
}

func getDelimiter(delimiter string) string {
	if delimiter == "" {
		return ","
	}
	return delimiter
}

func parseResourceID(r *http.Request) (int64, error) {
	idStr := mux.Vars(r)["id"]
	return strconv.ParseInt(idStr, 10, 64)
}

func readCSV(file io.Reader, delimiter string) ([][]string, error) {
	reader := csv.NewReader(file)
	reader.Comma = rune(delimiter[0])
	return reader.ReadAll()
}

func findColumnIndex(headers []string, columnName string) (int, error) {
	for i, header := range headers {
		trimHeader := trimBOM([]byte(header))
		if string(trimHeader) == columnName {
			return i, nil
		}
	}
	return -1, errors.New("specified column not found in CSV headers")
}

func mapData(records [][]string, mapColIndex int, resourceId int64) ([]*model.ResourceDisplay, model.AppError) {
	var mappedData []*model.ResourceDisplay

	for i, row := range records[1:] {
		resouce := model.NewResourceDisplay(row[mapColIndex], resourceId)

		if err := resouce.Parse(); err != nil {
			return nil, model.NewBadRequestError(
				fmt.Sprintf("cc_outbound_resource.validatePhoneNumber.%d", i),
				"number must contain letters (a-z, A-Z), numbers (0-9), and special characters: +,-, _, ., !, ~, *, ', (,)",
			)
		}

		mappedData = append(mappedData, resouce)
	}

	return mappedData, nil
}

func writeJSONResponse(w http.ResponseWriter, data any) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
