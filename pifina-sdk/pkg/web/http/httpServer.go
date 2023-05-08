package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/r3labs/sse/v2"
	"github.com/thushjandan/pifina/pkg/model"
)

type PifinaHttpServer struct {
	logger hclog.Logger
	server *http.Server
	sse    *sse.Server
}

func NewPifinaHttpServer(logger hclog.Logger) *PifinaHttpServer {
	return &PifinaHttpServer{
		logger: logger.Named("api"),
	}
}

func (s *PifinaHttpServer) StartWebServer(ctx context.Context, port string, keyFile string, certFile string, metricChannel chan []*model.MetricItem) {
	s.sse = sse.New()
	s.sse.CreateStream("metrics")
	// Create a new Mux and set the handler
	mux := http.NewServeMux()
	mux.HandleFunc("/events", s.sse.ServeHTTP)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	go s.ListenAndPublishMetrics(ctx, metricChannel)

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

func (s *PifinaHttpServer) ListenAndPublishMetrics(ctx context.Context, metricChannel chan []*model.MetricItem) {
	s.logger.Info("Starting http/2 sse server.")
	for {
		select {
		case metricList := <-metricChannel:
			s.logger.Trace("Got metrics", "metrics", metricList)
			s.sse.Publish("metrics", &sse.Event{
				Data: []byte("ping"),
			})
		case <-ctx.Done():
			s.logger.Info("Stopping sse server.")
			return
		}
	}
}
