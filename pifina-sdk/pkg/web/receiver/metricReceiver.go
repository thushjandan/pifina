package receiver

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/sink/protos/pifina/pifina"
	"google.golang.org/protobuf/proto"
)

type MetricReceiver struct {
	logger hclog.Logger
	conn   *net.UDPConn
}

func NewPifinaMetricReceiver(logger hclog.Logger) *MetricReceiver {
	return &MetricReceiver{
		logger: logger.Named("metric-receiver"),
	}
}

func (r *MetricReceiver) StartServer(ctx context.Context, port string) error {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}

	r.conn, err = net.ListenUDP("udp", serverAddr)
	r.logger.Info("Starting receiver", "port", port)
	// Runs the UDP server
	go func() {
		buf := make([]byte, 2048)
		for {
			// If termination signal has received, terminate udp server.
			if context.Cause(ctx) != nil {
				return
			}

			n, _, err := r.conn.ReadFromUDP(buf)
			if err != nil {
				// If termination signal has received, terminate udp server.
				if context.Cause(ctx) != nil {
					return
				}
				r.logger.Error("Error occured during reading from UDP packet", "err", err, "type")
				continue
			}

			protoTelemetryMsg := &pifina.PifinaTelemetryMessage{}
			err = proto.Unmarshal(buf[0:n], protoTelemetryMsg)
			if err != nil {
				r.logger.Error("Cannot decode protobuf message from UDP packet", "err", err)
				continue
			}
			r.logger.Debug("Successfully decoded protobuf telemetry message", "host", protoTelemetryMsg.SourceHost)
		}

	}()

	return nil
}

func (r *MetricReceiver) Shutdown() {
	r.logger.Info("Stopping metric receiver...")
	r.conn.Close()
}
