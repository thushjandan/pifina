// Copyright (c) 2023 Thushjandan Ponnudurai
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sink

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/model"
	"github.com/thushjandan/pifina/pkg/model/protos/pifina/pifina"
	"google.golang.org/protobuf/proto"
)

type Sink struct {
	logger         hclog.Logger
	pifinaEndpoint string
	hostType       pifina.PifinaHostTypes
	mySystemName   string
	groupId        uint32
}

// hostType needs to be one of the constants defined in metricItemModel file
// possible values: model.HOSTTYPE_TOFINO or model.HOSTTYPE_NIC
// group id used to group multiple probes in the frontend
func NewSink(logger hclog.Logger, hostType string, pifinaEndpoint string, groupId uint32) *Sink {
	logger = logger.Named("sink")
	hostname, err := os.Hostname()
	if err != nil {
		logger.Error("Cannot retrieve system hostname. setting system name to unknown")
		hostname = "unknown"
	}

	// Check host type parameter
	var pfHostType pifina.PifinaHostTypes
	switch hostType {
	case model.HOSTTYPE_TOFINO:
		pfHostType = pifina.PifinaHostTypes_TYPE_TOFINO
	case model.HOSTTYPE_NIC:
		pfHostType = pifina.PifinaHostTypes_TYPE_NIC
	default:
		pfHostType = pifina.PifinaHostTypes_TYPE_UNSPECIFIED
	}

	return &Sink{
		logger:         logger,
		pifinaEndpoint: pifinaEndpoint,
		hostType:       pfHostType,
		mySystemName:   hostname,
		groupId:        groupId,
	}
}

func (s *Sink) StartSink(ctx context.Context, wg *sync.WaitGroup, c chan *model.SinkEmitCommand) error {
	defer wg.Done()

	for {
		select {
		case batch := <-c:
			// Chunk metric slice in size of 20 items
			// Avoids UDP fragmentation => below 1460 bytes.
			metricChunks := chunkSlice(batch.Metrics, 20)
			for i := range metricChunks {
				var err error
				if batch.SourceSuffix != "" {
					err = s.emitWithSource(metricChunks[i], fmt.Sprintf("%s_%s", s.mySystemName, batch.SourceSuffix))
				} else {
					err = s.emit(metricChunks[i])
				}
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
func (s *Sink) emit(metrics []*model.MetricItem) error {
	return s.emitWithSource(metrics, s.mySystemName)
}

// Transforms the payload to protobuf and sends to pifina server
// Source can be modified by caller
func (s *Sink) emitWithSource(metrics []*model.MetricItem, sourceName string) error {
	protobufMetrics := model.ConvertMetricsToProtobuf(metrics)
	telemetryPayload := &pifina.PifinaTelemetryMessage{
		SourceHost: sourceName,
		HostType:   s.hostType,
		GroupId:    s.groupId,
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
	s.logger.Debug("Metrics have been sent to pifina server", "source", sourceName, "server", s.pifinaEndpoint)

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
