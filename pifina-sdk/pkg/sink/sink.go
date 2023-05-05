package sink

import (
	"context"
	"net"
	"os"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/dataplane/tofino/driver"
	"github.com/thushjandan/pifina/pkg/sink/protos/pifina/pifina"
)

type Sink struct {
	logger         hclog.Logger
	pifinaEndpoint string
	mySystemName   string
}

func NewSink(logger hclog.Logger, pifinaEndpoint string) *Sink {
	logger = logger.Named("sink")
	hostname, err := os.Hostname()
	if err != nil {
		logger.Error("Cannot retrieve system hostname. setting system name to unknown")
		hostname = "unknown"
	}
	return &Sink{
		logger:         logger,
		pifinaEndpoint: pifinaEndpoint,
		mySystemName:   hostname,
	}
}

func (s *Sink) StartSink(ctx context.Context, wg *sync.WaitGroup, c chan []*driver.MetricItem) error {
	defer wg.Done()

	for {
		select {
		case metrics := <-c:
			err := s.Emit(metrics)
			if err != nil {
				s.logger.Error("Error occured the transmission of the metrics", "error", err)
			}
		case <-ctx.Done():
			s.logger.Info("Stopping pifina sink...")
			return nil
		}
	}

}

// Transforms the payload to protobuf and sends to pifina server
func (s *Sink) Emit(metrics []*driver.MetricItem) error {
	protobufMetrics := ConvertMetricsToProtobuf(metrics)
	telemetryPayload := &pifina.PifinaTelemetryMessage{
		SourceHost: s.mySystemName,
		Metrics:    protobufMetrics,
	}

	// Convert to byte string
	s.logger.Debug("Marshalling metrics to protobuf")
	data, err := proto.Marshal(telemetryPayload)
	if err != nil {
		return err
	}

	// Resolve UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", s.pifinaEndpoint)
	if err != nil {
		return err
	}

	// Connect to Pifina Server
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}

	defer conn.Close()

	// Send metrics to server
	_, err = conn.Write([]byte(data))
	if err != nil {
		return err
	}
	s.logger.Debug("Metrics has been sent to pifina server", "server", s.pifinaEndpoint)

	return nil
}
