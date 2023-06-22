package sink

import (
	"context"
	"net"
	"os"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/model"
	"github.com/thushjandan/pifina/pkg/model/protos/pifina/pifina"
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

func (s *Sink) StartSink(ctx context.Context, wg *sync.WaitGroup, c chan []*model.MetricItem) error {
	defer wg.Done()

	for {
		select {
		case metrics := <-c:
			// Chunk metric slice in size of 20 items
			// Avoids UDP fragmentation => below 1460 bytes.
			metricChunks := chunkSlice(metrics, 20)
			for i := range metricChunks {
				err := s.Emit(metricChunks[i])
				if err != nil {
					s.logger.Error("Error occured the transmission of the metrics", "error", err)
				}
			}
		case <-ctx.Done():
			s.logger.Info("Stopping pifina sink...")
			return nil
		}
	}

}

// Transforms the payload to protobuf and sends to pifina server
func (s *Sink) Emit(metrics []*model.MetricItem) error {
	protobufMetrics := model.ConvertMetricsToProtobuf(metrics)
	telemetryPayload := &pifina.PifinaTelemetryMessage{
		SourceHost: s.mySystemName,
		Metrics:    protobufMetrics,
	}

	// Convert to byte string
	s.logger.Trace("Marshalling metrics to protobuf")
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

func chunkSlice(slice []*model.MetricItem, chunkSize int) [][]*model.MetricItem {
	var chunks [][]*model.MetricItem
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}
