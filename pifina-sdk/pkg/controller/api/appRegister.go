// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

import (
	"encoding/json"
	"net/http"

	"github.com/thushjandan/pifina/pkg/model"
)

func (s *ControllerApiServer) HandleAppRegisterReq(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getAppRegisterProbes(rw, r)
	case http.MethodPost:
		s.createAppRegisterProbe(rw, r)
	case http.MethodDelete:
		s.deleteAppRegisterProbe(rw, r)
	}
}

// Returns configured app registers to monitor
func (s *ControllerApiServer) getAppRegisterProbes(rw http.ResponseWriter, r *http.Request) {
	registers := s.ts.GetAppRegisterProbes()
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(registers)
}

// Returns the names all existing Registers
func (s *ControllerApiServer) GetAllAppRegisterNames(rw http.ResponseWriter, r *http.Request) {
	registers := s.ts.GetAllAppRegistersOnDevice()
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(registers)
}

func (s *ControllerApiServer) createAppRegisterProbe(rw http.ResponseWriter, r *http.Request) {
	var newEntry *model.AppRegister
	err := json.NewDecoder(r.Body).Decode(&newEntry)
	if err != nil {
		errorMessage := &model.ApiErrorMessage{Message: "Invalid json. Check your input", Code: http.StatusBadRequest}
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(errorMessage)
		return
	}
	if newEntry.Name == "" {
		errorMessage := &model.ApiErrorMessage{Message: "Invalid name. Check your input", Code: http.StatusBadRequest}
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(errorMessage)
		return
	}

	err = s.ts.AddAppRegisterProbe(newEntry)
	if err != nil {
		errorMessage := &model.ApiErrorMessage{Message: err.Error(), Code: http.StatusBadRequest}
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(errorMessage)
		return
	}

	rw.WriteHeader(http.StatusCreated)
}

func (s *ControllerApiServer) deleteAppRegisterProbe(rw http.ResponseWriter, r *http.Request) {
	var newEntry *model.AppRegister
	err := json.NewDecoder(r.Body).Decode(&newEntry)
	if err != nil {
		errorMessage := &model.ApiErrorMessage{Message: "Invalid json. Check your input", Code: http.StatusBadRequest}
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(errorMessage)
		return
	}
	if newEntry.Name == "" {
		errorMessage := &model.ApiErrorMessage{Message: "Invalid name. Check your input", Code: http.StatusBadRequest}
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(errorMessage)
		return
	}

	// Remove register from data collection
	s.ts.RemoveAppRegisterProbe(newEntry)

	rw.WriteHeader(http.StatusNoContent)
}
