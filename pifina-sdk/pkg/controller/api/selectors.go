package api

import (
	"encoding/json"
	"net/http"
)

func (s *ControllerApiServer) GetSelectors(rw http.ResponseWriter, r *http.Request) {
	matchSelectors := s.ts.GetTrafficSelectorCache()
	json.NewEncoder(rw).Encode(matchSelectors)
	rw.Header().Set("Content-Type", "application/json")
}

func (s *ControllerApiServer) AddNewSelector(rw http.ResponseWriter, r *http.Request) {
	s.logger.Debug("Adding a new selector on the control plance")
}

func (s *ControllerApiServer) HandleSelectorReq(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.GetSelectors(rw, r)
	case http.MethodPost:
		s.AddNewSelector(rw, r)
	case http.MethodOptions:
		rw.Header().Set("Allow", "GET, POST, OPTIONS")
		rw.WriteHeader(http.StatusNoContent)
	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}
