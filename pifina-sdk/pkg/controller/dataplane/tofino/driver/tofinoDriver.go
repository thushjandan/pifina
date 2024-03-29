// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package driver

import (
	"context"
	"math/rand"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/thushjandan/pifina/internal/dataplane/tofino/protos/bfruntime"
	"google.golang.org/grpc"
)

type TofinoDriver struct {
	logger               hclog.Logger
	p4Name               string
	isConnected          bool
	conn                 *grpc.ClientConn
	client               bfruntime.BfRuntimeClient
	lock                 sync.Mutex
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
	TABLE_TYPE_COUNTER                        = "Counter"
	TABLE_TYPE_MATCHACTION                    = "MatchAction_Direct"
	TABLE_TYPE_TM_CNT_IG                      = "TmCounterIgPort"
	TABLE_TYPE_TM_CNT_EG                      = "TmCounterEgPort"
	TABLE_NAME_PORT_INFO                      = "$PORT"
	PORT_NAME_INDEX_NAME                      = "$PORT_NAME"
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
func NewTofinoDriver(logger hclog.Logger, p4Name string) *TofinoDriver {
	return &TofinoDriver{
		logger:        logger.Named("tofinoDriver"),
		p4Name:        p4Name,
		isConnected:   false,
		clientId:      uint32(rand.Intn(100) + 1),
		probeTableMap: make(map[string]string),
	}
}
