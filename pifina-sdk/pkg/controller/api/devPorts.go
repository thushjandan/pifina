package api

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/thushjandan/pifina/pkg/model"
)

func (s *ControllerApiServer) GetAllAvailablePorts(rw http.ResponseWriter, r *http.Request) {
	ports := s.ts.GetAllAvailablePorts()
	sort.Strings(ports)
	json.NewEncoder(rw).Encode(ports)
	rw.WriteHeader(http.StatusOK)
}

func (s *ControllerApiServer) HandlePortsToMonitor(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.GetMonitoredPorts(rw, r)
	case http.MethodPost:
		s.AddPortToMonitor(rw, r)
	case http.MethodDelete:
		s.DeleteMonitoredPort(rw, r)
	}
}

func (s *ControllerApiServer) GetMonitoredPorts(rw http.ResponseWriter, r *http.Request) {
	ports := s.ts.GetMonitoredPorts()
	sort.Strings(ports)
	json.NewEncoder(rw).Encode(ports)
	rw.WriteHeader(http.StatusOK)
}

func (s *ControllerApiServer) AddPortToMonitor(rw http.ResponseWriter, r *http.Request) {
	var devPort *model.DevPort
	json.NewDecoder(r.Body).Decode(&devPort)
	s.ts.AddPortToMonitor(devPort.Name)
	rw.WriteHeader(http.StatusCreated)
}

func (s *ControllerApiServer) DeleteMonitoredPort(rw http.ResponseWriter, r *http.Request) {
	var devPort *model.DevPort
	json.NewDecoder(r.Body).Decode(&devPort)
	s.ts.RemovePortToMonitor(devPort.Name)
	rw.WriteHeader(http.StatusNoContent)
}
