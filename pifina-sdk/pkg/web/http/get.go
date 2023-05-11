package http

import (
	"encoding/json"
	"net/http"
)

func (s *PifinaHttpServer) GetEndpointsHandler(rw http.ResponseWriter, r *http.Request) {
	endpoints := s.ed.GetAll()
	json.NewEncoder(rw).Encode(endpoints)
	rw.Header().Set("Content-Type", "application/json")
}
