package receiver

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/pkg/model"
	"github.com/thushjandan/pifina/pkg/model/protos/pifina/pifina"
	"github.com/thushjandan/pifina/pkg/web/endpoints"
	"google.golang.org/protobuf/proto"
)

type MetricReceiver struct {
	logger hclog.Logger
	ed     *endpoints.PifinaEndpointDirectory
	conn   *net.UDPConn
}

func NewPifinaMetricReceiver(logger hclog.Logger, ed *endpoints.PifinaEndpointDirectory) *MetricReceiver {
	return &MetricReceiver{
		logger: logger.Named("metric-receiver"),
		ed:     ed,
	}
}

func (r *MetricReceiver) StartServer(ctx context.Context, port string, telemetryChannel chan *model.TelemetryMessage) error {
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

			n, clientAddr, err := r.conn.ReadFromUDP(buf)
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
			r.logger.Trace("Successfully decoded protobuf telemetry message", "host", protoTelemetryMsg.SourceHost)
			metricList := model.ConvertProtobufToMetrics(protoTelemetryMsg.Metrics)
			// Check host type
			var hostType string
			switch protoTelemetryMsg.HostType {
			case pifina.PifinaHostTypes_TYPE_TOFINO:
				hostType = model.HOSTTYPE_TOFINO
			case pifina.PifinaHostTypes_TYPE_NIC:
				hostType = model.HOSTTYPE_NIC
			default:
				// Skip this metric as it is unknown
				continue
			}

			telemetryMessage := &model.TelemetryMessage{
				Source:     protoTelemetryMsg.SourceHost,
				HostType:   hostType,
				GroupId:    protoTelemetryMsg.GroupId,
				MetricList: metricList,
			}
			if len(metricList) > 0 {
				r.ed.Set(protoTelemetryMsg.SourceHost, hostType, protoTelemetryMsg.GroupId, clientAddr.IP)
				telemetryChannel <- telemetryMessage
			}
		}
	}()

	return nil
}

func (r *MetricReceiver) Shutdown() {
	r.logger.Info("Stopping metric receiver...")
	r.conn.Close()
}
