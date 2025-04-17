package apis

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/webitel/engine/model"
)

func (api *API) InitDisplays() {
	api.Routes.Root.Handle("/displays/{id}", api.ApiHandlerTrustRequester(createDisplays)).Methods("POST")
}

func createDisplays(c *Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10MB limit
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	delimiter := r.FormValue("delimiter")
	if delimiter == "" {
		delimiter = "," // Default delimiter
	}

	mapCol := r.FormValue("map")
	reader := csv.NewReader(file)

	// Read the header row
	headers, err := reader.Read()
	if err != nil {
		http.Error(w, "Error reading CSV file", http.StatusInternalServerError)
	}

	// Find the index of the column by name
	var columnIndex int
	for i, header := range headers {
		if header == mapCol {
			columnIndex = i
			break
		}
	}

	resourceId, err := strconv.ParseInt(getIdFromRequest(r), 10, 64)
	if err != nil {
		http.Error(w, "Error resourceId", http.StatusBadRequest)
	}

	records, err := readCSV(file, delimiter)
	if err != nil {
		http.Error(w, "Error reading CSV file", http.StatusInternalServerError)
		return
	}

	mappedData, err := mapData(records, columnIndex)
	if err != nil {
		http.Error(w, "Error mapping data", http.StatusInternalServerError)
		return
	}

	displays, err := c.App.CreateOutboundResourceDisplays(r.Context(), resourceId, mappedData)
	if err != nil {
		http.Error(w, "Failed to create outbound resource displays", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(displays); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func readCSV(file io.Reader, delimiter string) ([][]string, error) {
	reader := csv.NewReader(file)
	reader.Comma = rune(delimiter[0])
	return reader.ReadAll()
}

func mapData(records [][]string, mapColIndex int) ([]*model.ResourceDisplay, error) {
	var mappedData []*model.ResourceDisplay

	for _, row := range records[1:] {
		resourceId, err := strconv.ParseInt(row[0], 10, 64) //TODO: Add multiple resources ID's insert
		if err != nil {
			return nil, err
		}
		display := row[mapColIndex]
		mappedData = append(mappedData, &model.ResourceDisplay{
			ResourceId: resourceId,
			Display:    display,
		})
	}

	return mappedData, nil
}
