package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/r3labs/sse/v2"
	"github.com/thushjandan/pifina/pkg/model"
	"github.com/thushjandan/pifina/pkg/web/endpoints"
)

type PifinaHttpServer struct {
	logger hclog.Logger
	ed     *endpoints.PifinaEndpointDirectory
	server *http.Server
	sse    *sse.Server
}

func NewPifinaHttpServer(logger hclog.Logger, ed *endpoints.PifinaEndpointDirectory) *PifinaHttpServer {
	return &PifinaHttpServer{
		logger: logger.Named("api"),
		ed:     ed,
	}
}

func (s *PifinaHttpServer) StartWebServer(ctx context.Context, port string, keyFile string, certFile string, telemetryChannel chan *model.TelemetryMessage) {
	s.sse = sse.New()
	s.sse.EventTTL = 15 * time.Second
	s.sse.CreateStream("metrics")
	// Create a new Mux and set the handler
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/events", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Access-Control-Allow-Origin", "*")
		s.sse.ServeHTTP(rw, r)
	})
	mux.HandleFunc("/api/v1/endpoints", s.GetEndpointsHandler)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	go s.ListenAndPublishMetrics(ctx, telemetryChannel)

	s.logger.Info("Starting http/2 TLS server")
	if err := s.server.ListenAndServeTLS(certFile, keyFile); err != http.ErrServerClosed {
		s.logger.Error("Cannot start http server", "err", err)
	}
}

func (s *PifinaHttpServer) Shutdown() {
	s.logger.Info("Stopping API server")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(timeoutCtx); err != nil {
		s.logger.Error("Webserver shutdown failed", "err", err)
	}
}

func (s *PifinaHttpServer) ListenAndPublishMetrics(ctx context.Context, telemetryChannel chan *model.TelemetryMessage) {
	s.logger.Info("Starting http/2 sse server.")
	for {
		select {
		case telemetryItem := <-telemetryChannel:
			if !s.sse.StreamExists(telemetryItem.Source) {
				s.sse.CreateStream(telemetryItem.Source)
			}
			jsonPayload, err := json.Marshal(telemetryItem.MetricList)
			if err != nil {
				continue
			}

			s.sse.Publish(telemetryItem.Source, &sse.Event{
				Data: jsonPayload,
			})
		case <-ctx.Done():
			s.logger.Info("Stopping sse server.")
			return
		}
	}
}
