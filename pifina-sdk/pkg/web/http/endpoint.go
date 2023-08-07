// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package http

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/thushjandan/pifina/pkg/model"
)

func (s *PifinaHttpServer) GetEndpointsHandler(rw http.ResponseWriter, r *http.Request) {
	endpoints := s.ed.GetAll()
	json.NewEncoder(rw).Encode(endpoints)
	rw.Header().Set("Content-Type", "application/json")
}

// Update controller endpoint configuration like address or port
func (s *PifinaHttpServer) UpdateEndpointsHandler(rw http.ResponseWriter, r *http.Request) {
	var entry *model.ApiEndpointModel
	err := json.NewDecoder(r.Body).Decode(&entry)
	if err != nil {
		errorMessage := &model.ApiErrorMessage{Message: "Invalid json. Check your input", Code: http.StatusBadRequest}
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(errorMessage)
		return
	}
	endpointIp := net.ParseIP(entry.Address)
	if endpointIp == nil || entry.Endpoint == "" || entry.Port == 0 {
		errorMessage := &model.ApiErrorMessage{Message: "Invalid payload. Check your input", Code: http.StatusBadRequest}
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(errorMessage)
		return
	}

	ok := s.ed.Update(entry.Endpoint, endpointIp, entry.Port)
	if !ok {
		errorMessage := &model.ApiErrorMessage{Message: "Endpoint not found. Check your input", Code: http.StatusNotFound}
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(errorMessage)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *PifinaHttpServer) HandleEndpointRequest(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.GetEndpointsHandler(rw, r)
	case http.MethodPut:
		s.UpdateEndpointsHandler(rw, r)
	case http.MethodOptions:
		rw.Header().Set("Allow", "GET, PUT, OPTIONS")
		rw.WriteHeader(http.StatusNoContent)
	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}
