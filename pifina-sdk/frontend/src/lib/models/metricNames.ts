import { tickFormat } from "d3";
import type { PIFINA_CHART_CONFIG, PIFINA_CHART_CONF_ITEM } from "./dashboardConfigModel";
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
export const PROBE_NEO_TPT_MTT_L0_MISS = "Level 0 MTT Cache Miss"
export const PROBE_NEO_TPT_MTT_L1_MISS = "Level 1 MTT Cache Miss"
export const PROBE_NEO_TPT_MPT_L0_MISS = "Level 0 MPT Cache Miss"
export const PROBE_NEO_TPT_MPT_L1_MISS = "Level 1 MPT Cache Miss"
export const PROBE_NEO_PCI_BP = "PCIe Internal Back Pressure"
export const PROBE_NEO_ICM_MISS = "ICM Cache Miss"
export const PROBE_NEO_RX_FULL_0 = "RX Packet Buffer Full Port 0"
export const PROBE_NEO_RX_FULL_1 = "RX Packet Buffer Full Port 1"
export const PROBE_NEO_WQE_MISS = "Receive WQE Cache Miss"
export const PROBE_NEO_TX_BW = "TX BandWidth"
export const PROBE_NEO_RX_BW = "RX BandWidth"
export const PROBE_NEO_TX_PKT = "TX Packet Rate"
export const PROBE_NEO_RX_PKT = "RX Packet Rate"
export const PROBE_NEO_PCI_IN_BW = "PCIe Inbound BW Utilization"
export const PROBE_NEO_PCI_OUT_BW = "PCIe Outbound BW Utilization"
export const PROBE_ETHTOOL_RX_DISCARD = "rx_discards_phy"
export const PROBE_ETHTOOL_TX_DISCARD = "tx_discards_phy"
export const PROBE_ETHTOOL_RX_PAUSE = "rx_pause_ctrl_phy"
export const PROBE_ETHTOOL_TX_PAUSE = "tx_pause_ctrl_phy"
export const PROBE_ETHTOOL_RX_OOB = "rx_out_of_buffer"

export const Y_AXIS_NAME_BYTE_RATE = "byte/sec"
export const Y_AXIS_NAME_PKT_RATE = "pkts/sec"
export const Y_AXIS_NAME_TIME_MS = "ms"
export const Y_AXIS_NAME_TIME_SEC = "sec"
export const Y_AXIS_NAME_PKT_COUNT = "pkts"
export const Y_AXIS_NAME_CELL_COUNT = "cells"
export const Y_AXIS_NAME_EVENTS_COUNT = "events"
export const Y_AXIS_NAME_EVENTS_RATE = "events/sec"
export const Y_AXIS_NAME_CYCLES_RATE = "cycles/sec"
export const Y_AXIS_NAME_GIGABYTE_RATE = "Gb/sec"

export const PIFINA_DEFAULT_PROBES = [
    PROBE_INGRESS_MATCH_CNT_BYTE,
    PROBE_INGRESS_MATCH_CNT_PKT,
    PROBE_INGRESS_START_HDR_BYTE,
    PROBE_INGRESS_END_HDR_BYTE,
    PROBE_EGRESS_START_CNT_BYTE,
    PROBE_EGRESS_START_CNT_PKTS,
    PROBE_EGRESS_END_CNT_BYTE
]
