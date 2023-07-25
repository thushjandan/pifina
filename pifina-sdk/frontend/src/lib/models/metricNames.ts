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
export const Y_AXIS_NAME_BYTE_RATE = "byte/sec"
export const Y_AXIS_NAME_PKT_RATE = "pkts/sec"
export const Y_AXIS_NAME_TIME_MS = "ms"
export const Y_AXIS_NAME_PKT_COUNT = "pkts"
export const Y_AXIS_NAME_CELL_COUNT = "cells"

export const PIFINA_DEFAULT_PROBES = [
    PROBE_INGRESS_MATCH_CNT_BYTE,
    PROBE_INGRESS_MATCH_CNT_PKT,
    PROBE_INGRESS_START_HDR_BYTE,
    PROBE_INGRESS_END_HDR_BYTE,
    PROBE_EGRESS_START_CNT_BYTE,
    PROBE_EGRESS_START_CNT_PKTS,
    PROBE_EGRESS_END_CNT_BYTE
]

export const PIFINA_PROBE_CHART_CFG = {
    [PROBE_INGRESS_MATCH_CNT_BYTE]: {
        yAxisName: Y_AXIS_NAME_BYTE_RATE,
        title: "Ingress byte counter"
    },
    [PROBE_INGRESS_MATCH_CNT_PKT]: {
        yAxisName: Y_AXIS_NAME_PKT_RATE,
        title: "Ingress packet counter"
    },
    [PROBE_INGRESS_START_HDR_BYTE]: {
        yAxisName: Y_AXIS_NAME_BYTE_RATE,
        title: "Start ingress header size counter"
    },
    [PROBE_INGRESS_END_HDR_BYTE]: {
        yAxisName: Y_AXIS_NAME_BYTE_RATE,
        title: "End ingress header size counter"
    },
    [PROBE_EGRESS_START_CNT_BYTE]: {
        yAxisName: Y_AXIS_NAME_BYTE_RATE,
        title: "Start egress byte counter"
    },
    [PROBE_EGRESS_START_CNT_PKTS]: {
        yAxisName: Y_AXIS_NAME_PKT_RATE,
        title: "Egress packet counter"
    },
    [PROBE_EGRESS_END_CNT_BYTE]: {
        yAxisName: Y_AXIS_NAME_BYTE_RATE,
        title: "End egress byte counter"
    },
    [PROBE_INGRESS_JITTER]: {
        yAxisName: Y_AXIS_NAME_TIME_MS,
        title: "Ingress inter packet arrival average rate"
    },
    [PROBE_TM_INGRESS_DROP_PKT]: {
        yAxisName: Y_AXIS_NAME_PKT_COUNT,
        title: "Ingress packet drops from TM perspective"
    },
    [PROBE_TM_EGRESS_DROP_PKT]: {
        yAxisName: Y_AXIS_NAME_PKT_COUNT,
        title: "Egress packet drops from TM perspective"
    },
    [PROBE_TM_INRESS_USAGE_CELLS]: {
        yAxisName: Y_AXIS_NAME_CELL_COUNT,
        title: "Port usage count in terms of number of memory cells usage from TM ingress perspective"
    },
    [PROBE_TM_ERESS_USAGE_CELLS]: {
        yAxisName: Y_AXIS_NAME_CELL_COUNT,
        title: "Port usage count in terms of number of memory cells usage from TM egress perspective"
    },
    [PROBE_TM_PIPE_TOTAL_BUF_DROP]: {
        yAxisName: Y_AXIS_NAME_PKT_COUNT,
        title: "Number of packets which were dropped because of buffer full condition"
    },
    [PROBE_TM_PIPE_IG_FULL_BUF]: {
        yAxisName: Y_AXIS_NAME_PKT_COUNT,
        title: "The number of packets which were dropped because of buffer full condition on ingress side"
    },
    [PROBE_TM_PIPE_EG_DROP_PKT]: {
        yAxisName: Y_AXIS_NAME_PKT_COUNT,
        title: "The total number of packets which were dropped on egress side"
    },
}

export const PIFINA_DEFAULT_PROBE_CHART_ORDER = [PROBE_INGRESS_MATCH_CNT_BYTE, [PROBE_EGRESS_START_CNT_BYTE, PROBE_EGRESS_END_CNT_BYTE], [PROBE_INGRESS_MATCH_CNT_PKT, PROBE_EGRESS_START_CNT_PKTS], [PROBE_INGRESS_START_HDR_BYTE, PROBE_INGRESS_END_HDR_BYTE], PROBE_INGRESS_JITTER]