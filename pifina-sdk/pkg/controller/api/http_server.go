package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/controller/collector"
)

type ControllerApiServer struct {
	logger    hclog.Logger
	server    *http.Server
	collector *collector.MetricCollector
}

func NewPifinaHttpServer(logger hclog.Logger, c *collector.MetricCollector) *ControllerApiServer {
	return &ControllerApiServer{
		logger:    logger.Named("api"),
		collector: c,
	}
}

func (s *ControllerApiServer) StartWebServer(ctx context.Context, port string) {
	// Create a new Mux and set the handler
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/selectors", s.GetSelectors)
	mux.HandleFunc("/api/v1/selectors", s.AddNewSelector)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	s.logger.Info("Starting http server")
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
