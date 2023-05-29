package api

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/thushjandan/pifina/pkg/model"
)

func (s *ControllerApiServer) GetAllAvailablePorts(rw http.ResponseWriter, r *http.Request) {
	ports := s.ts.GetAllAvailablePorts()
	sort.Slice(ports, func(i, j int) bool { return ports[i].Name < ports[j].Name })
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(ports)
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
	transformedPorts := make([]*model.DevPort, 0, len(ports))
	for i := range ports {
		transformedPorts = append(transformedPorts, &model.DevPort{Name: ports[i]})
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(transformedPorts)
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
	//TODO Delete from buffer pool
	//s.bp.RemoveMetric(devPort.Name, newEntry.Index, model.METRIC_EXT_VALUE)
	rw.WriteHeader(http.StatusNoContent)
}
