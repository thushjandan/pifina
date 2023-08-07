// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package api

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/controller/bufferpool"
	"github.com/thushjandan/pifina/pkg/controller/trafficselector"
)

type ControllerApiServer struct {
	logger hclog.Logger
	port   string
	server *http.Server
	ts     *trafficselector.TrafficSelector
	bp     *bufferpool.Bufferpool
}

func NewControllerApiServer(logger hclog.Logger, port string, ts *trafficselector.TrafficSelector, bp *bufferpool.Bufferpool) *ControllerApiServer {
	return &ControllerApiServer{
		logger: logger.Named("api"),
		ts:     ts,
		bp:     bp,
		port:   port,
	}
}

func (s *ControllerApiServer) StartWebServer(ctx context.Context) {
	// Create a new Mux and set the handler
	mux := http.NewServeMux()
	mux.Handle("/api/v1/selectors", middlewareCORS(http.HandlerFunc(s.HandleSelectorReq)))
	mux.Handle("/api/v1/schema", middlewareCORS(http.HandlerFunc(s.GetSelectorSchema)))
	mux.Handle("/api/v1/app-registers", middlewareCORS(http.HandlerFunc(s.HandleAppRegisterReq)))
	mux.Handle("/api/v1/app-registers/available", middlewareCORS(http.HandlerFunc(s.GetAllAppRegisterNames)))
	mux.Handle("/api/v1/ports", middlewareCORS(http.HandlerFunc(s.HandlePortsToMonitor)))
	mux.Handle("/api/v1/ports/available", middlewareCORS(http.HandlerFunc(s.GetAllAvailablePorts)))

	s.server = &http.Server{
		Addr:    s.port,
		Handler: mux,
	}

	s.logger.Info("Starting API server")
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		s.logger.Error("Cannot start http server", "err", err)
	}
}

func (s *ControllerApiServer) Shutdown() {
	s.logger.Info("Stopping API server")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(timeoutCtx); err != nil {
		s.logger.Error("Webserver shutdown failed", "err", err)
	}
}

func middlewareCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.Header().Set("Access-Control-Allow-Origin", "*")
		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
		rw.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		next.ServeHTTP(rw, r)
	})
}
