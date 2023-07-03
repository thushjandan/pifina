package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/r3labs/sse/v2"
	"github.com/thushjandan/pifina"
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
	assets, _ := pifina.Assets()
	fs := http.FileServer(http.FS(assets))

	s.sse = sse.New()
	// Disable Replay feature from SSE
	s.sse.AutoReplay = false
	// Add CORS header => allows all origins
	s.sse.Headers["Access-Control-Allow-Origin"] = "*"

	// Create a new Mux and set the handler
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/events", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			// Received Browser Disconnection
			s.logger.Info("New client has connected")
			<-r.Context().Done()
			s.logger.Info("a client has disconnected")
			return
		}()

		s.sse.ServeHTTP(w, r)
	})
	mux.HandleFunc("/api/v1/endpoints", s.GetEndpointsHandler)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			f, err := assets.Open(strings.TrimPrefix(path.Clean(r.URL.Path), "/"))
			if err == nil {
				defer f.Close()
			}
			if os.IsNotExist(err) {
				r.URL.Path = "/"
			}
		}
		fs.ServeHTTP(w, r)
	})

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

	// Shutdown SSE Server. Disconnect all clients
	s.sse.Close()

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

			s.sse.TryPublish(telemetryItem.Source, &sse.Event{
				Data: jsonPayload,
			})
		case <-ctx.Done():
			s.logger.Info("Stopping sse server.")
			return
		}
	}
}
