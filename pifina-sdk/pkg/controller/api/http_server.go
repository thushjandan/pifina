package api

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/controller/trafficselector"
)

type ControllerApiServer struct {
	logger hclog.Logger
	port   string
	server *http.Server
	ts     *trafficselector.TrafficSelector
}

func NewControllerApiServer(logger hclog.Logger, port string, ts *trafficselector.TrafficSelector) *ControllerApiServer {
	return &ControllerApiServer{
		logger: logger.Named("api"),
		ts:     ts,
		port:   port,
	}
}

func (s *ControllerApiServer) StartWebServer(ctx context.Context) {
	// Create a new Mux and set the handler
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/selectors", s.HandleSelectorReq)

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
