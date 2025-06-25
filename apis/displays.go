package apis

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/webitel/engine/model"
)

func (api *API) InitDisplays() {
	api.Routes.Root.Handle("/api/displays/{id}", api.ApiHandlerTrustRequester(createDisplays)).Methods("POST")
}

func createDisplays(c *Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	delimiter := getDelimiter(r.FormValue("delimiter"))
	mapCol := r.FormValue("map")

	resourceId, err := parseResourceID(r)
	if err != nil {
		http.Error(w, "Invalid resource ID", http.StatusBadRequest)
		return
	}

	records, err := readCSV(file, delimiter)
	if err != nil {
		http.Error(w, "Error reading CSV file", http.StatusInternalServerError)
		return
	}

	if len(records) == 0 {
		http.Error(w, "CSV file is empty", http.StatusBadRequest)
		return
	}

	columnIndex, err := findColumnIndex(records[0], mapCol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mappedData, err := mapData(records, columnIndex, resourceId)
	if err != nil {
		http.Error(w, "Error mapping data", http.StatusInternalServerError)
		return
	}

	displays, err := c.App.CreateOutboundResourceDisplays(r.Context(), resourceId, mappedData)
	if err != nil {
		http.Error(w, "Failed to create outbound resource displays", http.StatusInternalServerError)
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
		if header == columnName {
			return i, nil
		}
	}
	return -1, errors.New("specified column not found in CSV headers")
}

func mapData(records [][]string, mapColIndex int, resourceId int64) ([]*model.ResourceDisplay, error) {
	var mappedData []*model.ResourceDisplay

	for _, row := range records[1:] {
		mappedData = append(mappedData, &model.ResourceDisplay{
			ResourceId: resourceId,
			Display:    row[mapColIndex],
		})
	}

	return mappedData, nil
}

func writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
