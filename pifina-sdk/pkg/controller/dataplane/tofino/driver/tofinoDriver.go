package driver

import (
	"context"
	"math/rand"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
	"google.golang.org/grpc"
)

type TofinoDriver struct {
	logger               hclog.Logger
	isConnected          bool
	conn                 *grpc.ClientConn
	client               bfruntime.BfRuntimeClient
	streamChannel        bfruntime.BfRuntime_StreamChannelClient
	ctx                  context.Context
	cancel               context.CancelFunc
	clientId             uint32
	P4Tables             []Table
	NonP4Tables          []Table
	indexP4Tables        map[string]int
	indexByIdP4Tables    map[uint32]int
	indexNonP4Tables     map[string]int
	indexByIdNonP4Tables map[uint32]int
	portCache            map[string][]byte
	probeTableMap        map[string]string
	extraProbeNameCache  []string
}

const (
	TOFINO_PIPE_ID                            = 0xffff // Target specific pipelines. 0xffff => ALL PIPELINES
	PROBE_INGRESS_MATCH_CNT                   = "PF_INGRESS_MATCH_CNT"
	PROBE_INGRESS_START_HDR_SIZE              = "PF_INGRESS_START_HDR_SIZE"
	PROBE_INGRESS_END_HDR_SIZE                = "PF_INGRESS_END_HDR_SIZE"
	PROBE_EGRESS_START_CNT                    = "PF_EGRESS_START_CNT"
	PROBE_EGRESS_END_CNT                      = "PF_EGRESS_END_CNT"
	PROBE_INGRESS_MATCH_ACTION_NAME           = "pf_start_ingress_measure"
	PROBE_INGRESS_MATCH_ACTION_NAME_SESSIONID = "sessionId"
	COUNTER_INDEX_KEY_NAME                    = "$COUNTER_INDEX"
	REGISTER_INDEX_KEY_NAME                   = "$REGISTER_INDEX"
	LPF_INDEX_KEY_NAME                        = "$LPF_INDEX"
	COUNTER_SPEC_BYTES                        = "$COUNTER_SPEC_BYTES"
	COUNTER_SPEC_PKTS                         = "$COUNTER_SPEC_PKTS"
	LPF_SPEC_TYPE                             = "$LPF_SPEC_TYPE"
	LPF_GAIN_TIME                             = "$LPF_SPEC_GAIN_TIME_CONSTANT_NS"
	LPF_DECAY_TIME                            = "$LPF_SPEC_DECAY_TIME_CONSTANT_NS"
	LPF_SCALE_DOWN_FACTOR                     = "$LPF_SPEC_OUT_SCALE_DOWN_FACTOR"
	TABLE_TYPE_REGISTER                       = "Register"
	TABLE_NAME_PORT_INFO                      = "$PORT_STR_INFO"
	TABLE_NAME_TM_CNT_IG                      = "tf2.tm.counter.ig_port"
	TABLE_NAME_TM_CNT_EG                      = "tf2.tm.counter.eg_port"
	TABLE_NAME_TM_CNT_PIPE                    = "tf2.tm.counter.pipe"
	DEV_PORT_KEY_NAME                         = "dev_port"
	PROBE_EXTRA_PREFIX                        = "PF_EXTRA"
	PROBE_INGRESS_JITTER_LPF                  = "PF_INGRESS_JITTER_LPF"
	PROBE_INGRESS_JITTER_REGISTER             = "PF_INGRESS_JITTER_AVG"
)

var PROBE_TABLES = []string{PROBE_INGRESS_MATCH_CNT, PROBE_INGRESS_START_HDR_SIZE, PROBE_INGRESS_END_HDR_SIZE, PROBE_EGRESS_START_CNT, PROBE_EGRESS_END_CNT, PROBE_INGRESS_JITTER_LPF, PROBE_INGRESS_JITTER_REGISTER}

// Creates new Tofino driver object
func NewTofinoDriver(logger hclog.Logger) *TofinoDriver {
	return &TofinoDriver{
		logger:        logger.Named("tofinoDriver"),
		isConnected:   false,
		clientId:      uint32(rand.Intn(100)),
		probeTableMap: make(map[string]string),
	}
}
