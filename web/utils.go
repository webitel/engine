package web

import "net/http"

func WriteJSON(w http.ResponseWriter, b []byte) (int, error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	return w.Write(b)
}
