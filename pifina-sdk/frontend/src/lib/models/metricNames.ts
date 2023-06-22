import { MetricTypes, PifinaMetricName } from "./metricTypes";

export const PROBE_INGRESS_MATCH_CNT_BYTE = `${PifinaMetricName.INGRESS_MATCH_CNT}${MetricTypes.BYTES}`
export const PROBE_INGRESS_MATCH_CNT_PKT = `${PifinaMetricName.INGRESS_MATCH_CNT}${MetricTypes.PKTS}`
export const PROBE_INGRESS_START_HDR_BYTE = `${PifinaMetricName.INGRESS_START_HDR}${MetricTypes.BYTES}`
export const PROBE_INGRESS_END_HDR_BYTE = `${PifinaMetricName.INGRESS_END_HDR}${MetricTypes.BYTES}`
export const PROBE_EGRESS_START_CNT_BYTE = `${PifinaMetricName.EGRESS_START_CNT}${MetricTypes.BYTES}`
export const PROBE_EGRESS_START_CNT_PKTS = `${PifinaMetricName.EGRESS_START_CNT}${MetricTypes.PKTS}`
export const PROBE_EGRESS_END_CNT_BYTE = `${PifinaMetricName.EGRESS_END_CNT}${MetricTypes.BYTES}`
export const PROBE_INGRESS_JITTER = `${PifinaMetricName.INGRESS_JITTER_AVG}${MetricTypes.EXT_VALUE}`
export const PROBE_EXTRA_PREFIX = "PF_EXTRA"
export const PROBE_TM_INGRESS_DROP_PKT = `PF_TM_ig_port_drop_count_packets`;
export const PROBE_TM_EGRESS_DROP_PKT = `PF_TM_eg_port_drop_count_packets`;
export const PROBE_TM_INRESS_USAGE_CELLS = `PF_TM_ig_port_usage_cells`;
export const PROBE_TM_ERESS_USAGE_CELLS = `PF_TM_eg_port_usage_cells`;
export const PROBE_TM_PIPE_TOTAL_BUF_DROP = `PF_TM_pipe_total_buffer_full_drop_packets`;
export const PROBE_TM_PIPE_IG_FULL_BUF = `PF_TM_pipe_ig_buf_full_drop_packets`;
export const PROBE_TM_PIPE_EG_DROP_PKT = `PF_TM_pipe_eg_total_drop_packets`;

export const PIFINA_DEFAULT_PROBES = [
    PROBE_INGRESS_MATCH_CNT_BYTE,
    PROBE_INGRESS_MATCH_CNT_PKT,
    PROBE_INGRESS_START_HDR_BYTE,
    PROBE_INGRESS_END_HDR_BYTE,
    PROBE_EGRESS_START_CNT_BYTE,
    PROBE_EGRESS_START_CNT_PKTS,
    PROBE_EGRESS_END_CNT_BYTE
]