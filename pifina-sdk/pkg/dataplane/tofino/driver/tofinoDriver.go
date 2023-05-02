package driver

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
	"google.golang.org/grpc"
)

type TofinoDriver struct {
	logger           hclog.Logger
	isConnected      bool
	conn             *grpc.ClientConn
	client           bfruntime.BfRuntimeClient
	streamChannel    bfruntime.BfRuntime_StreamChannelClient
	ctx              context.Context
	cancel           context.CancelFunc
	clientId         uint32
	P4Tables         []Table
	NonP4Tables      []Table
	indexP4Tables    map[string]int
	indexNonP4Tables map[string]int
	portCache        map[string][]byte
	probeTableMap    map[string]string
}

type MetricItem struct {
	SessionId uint32
	Type      string
	Value     uint64
}

type ErrNameNotFound struct {
	Entity string
	Msg    string
}

func (e *ErrNameNotFound) Error() string {
	return fmt.Sprintf("%s - Entity: %s", e.Msg, e.Entity)
}

const (
	PROBE_INGRESS_MATCH_CNT                   = "PF_INGRESS_MATCH_CNT"
	PROBE_INGRESS_START_HDR_SIZE              = "PF_INGRESS_START_HDR_SIZE"
	PROBE_INGRESS_END_HDR_SIZE                = "PF_INGRESS_END_HDR_SIZE"
	PROBE_EGRESS_START_CNT                    = "PF_EGRESS_START_CNT"
	PROBE_EGRESS_END_CNT                      = "PF_EGRESS_END_CNT"
	PROBE_INGRESS_MATCH_ACTION_NAME           = "pf_start_ingress_measure"
	PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID = "sessionId"
	REGISTER_KEY_NAME                         = "$COUNTER_INDEX"
)

var PROBE_TABLES = []string{PROBE_INGRESS_MATCH_CNT, PROBE_INGRESS_START_HDR_SIZE, PROBE_INGRESS_END_HDR_SIZE, PROBE_EGRESS_START_CNT, PROBE_EGRESS_END_CNT}

// Creates new Tofino driver object
func NewTofinoDriver(logger hclog.Logger) *TofinoDriver {
	return &TofinoDriver{
		logger:        logger.Named("tofinoDriver"),
		isConnected:   false,
		clientId:      uint32(rand.Intn(100)),
		probeTableMap: make(map[string]string),
	}
}
