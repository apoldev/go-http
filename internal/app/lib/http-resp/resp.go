package httpresp

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, data interface{}, code int) {
	bytes, err := json.Marshal(&data)
	if err != nil {
		Error(w, "invalid json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(bytes) //nolint:errcheck // ignore
}

func Error(w http.ResponseWriter, s string, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	w.Write([]byte(s)) //nolint:errcheck // ignore
}
