package http

import (
	"encoding/json"
	"net/http"
)

//ResponseSuccess ...
func ResponseSuccess(resp http.ResponseWriter, req *http.Request, data interface{}) error {
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
	return json.NewEncoder(resp).Encode(data)
}

//ResponseError ...
func ResponseError(resp http.ResponseWriter, req *http.Request, errorMessage string) error {
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusInternalServerError)
	return json.NewEncoder(resp).Encode(map[string]interface{}{"error": errorMessage})
}
