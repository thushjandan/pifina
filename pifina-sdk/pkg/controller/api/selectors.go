package api

import (
	"encoding/json"
	"net/http"

	"github.com/thushjandan/pifina/pkg/model"
)

func (s *ControllerApiServer) GetSelectors(rw http.ResponseWriter, r *http.Request) {
	matchSelectors := s.ts.GetTrafficSelectorCache()
	json.NewEncoder(rw).Encode(matchSelectors)
	rw.Header().Set("Content-Type", "application/json")
}

func (s *ControllerApiServer) AddNewSelector(rw http.ResponseWriter, r *http.Request) {
	var matchSelectorEntry model.MatchSelectorEntry

	err := json.NewDecoder(r.Body).Decode(&matchSelectorEntry)
	if err != nil {
		s.logger.Warn("Invalid request body for AddNewSelector API request", "err", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	err = s.ts.AddTrafficSelectorRule(&matchSelectorEntry)
	if err != nil {
		s.logger.Error("Adding new selector rule failed", "err", err)
		rw.Header().Set("Content-Type", "application/json")
		errorMessage := &model.ApiErrorMessage{Message: err.Error(), Code: http.StatusInternalServerError}
		rw.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rw).Encode(errorMessage)
		return
	}

	rw.WriteHeader(http.StatusCreated)
}

func (s *ControllerApiServer) RemoveSelector(rw http.ResponseWriter, r *http.Request) {
	var matchSelectorEntry model.MatchSelectorEntry

	err := json.NewDecoder(r.Body).Decode(&matchSelectorEntry)
	if err != nil {
		s.logger.Warn("Invalid request body for AddNewSelector API request", "err", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	err = s.ts.RemoveTrafficSelectorRule(&matchSelectorEntry)
	if err != nil {
		s.logger.Error("Removing selector rule failed", "err", err)
		rw.Header().Set("Content-Type", "application/json")
		errorMessage := &model.ApiErrorMessage{Message: err.Error(), Code: http.StatusInternalServerError}
		rw.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rw).Encode(errorMessage)
		return
	}

	rw.WriteHeader(http.StatusNoContent)

}

func (s *ControllerApiServer) HandleSelectorReq(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.GetSelectors(rw, r)
	case http.MethodPost:
		s.AddNewSelector(rw, r)
	case http.MethodDelete:
		s.RemoveSelector(rw, r)
	case http.MethodOptions:
		rw.Header().Set("Allow", "GET, POST, OPTIONS")
		rw.WriteHeader(http.StatusNoContent)
	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}
