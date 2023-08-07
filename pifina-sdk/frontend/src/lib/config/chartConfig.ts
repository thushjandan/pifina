// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import type { PIFINA_CHART_CONFIG, PIFINA_CHART_CONF_ITEM } from "$lib/models/dashboardConfigModel";
import * as pb from "$lib/models/metricNames";

export const PIFINA_PROBE_CHART_CFG: PIFINA_CHART_CONFIG = {
    [pb.PROBE_INGRESS_MATCH_CNT_BYTE]: {
        yAxisName: pb.Y_AXIS_NAME_BYTE_RATE,
        title: "Ingress byte counter"
    },
    [pb.PROBE_INGRESS_MATCH_CNT_PKT]: {
        yAxisName: pb.Y_AXIS_NAME_PKT_RATE,
        title: "Ingress packet counter"
    },
    [pb.PROBE_INGRESS_START_HDR_BYTE]: {
        yAxisName: pb.Y_AXIS_NAME_BYTE_RATE,
        title: "Start ingress header size counter"
    },
    [pb.PROBE_INGRESS_END_HDR_BYTE]: {
        yAxisName: pb.Y_AXIS_NAME_BYTE_RATE,
        title: "End ingress header size counter"
    },
    [pb.PROBE_EGRESS_START_CNT_BYTE]: {
        yAxisName: pb.Y_AXIS_NAME_BYTE_RATE,
        title: "Start egress byte counter"
    },
    [pb.PROBE_EGRESS_START_CNT_PKTS]: {
        yAxisName: pb.Y_AXIS_NAME_PKT_RATE,
        title: "Egress packet counter"
    },
    [pb.PROBE_EGRESS_END_CNT_BYTE]: {
        yAxisName: pb.Y_AXIS_NAME_BYTE_RATE,
        title: "End egress byte counter"
    },
    [pb.PROBE_INGRESS_JITTER]: {
        yAxisName: pb.Y_AXIS_NAME_TIME_SEC,
        title: "Ingress inter packet arrival average rate",
        tickFormat: "s" // see for format options: https://github.com/d3/d3-format#api-reference
    },
    [pb.PROBE_TM_INGRESS_DROP_PKT]: {
        yAxisName: pb.Y_AXIS_NAME_PKT_COUNT,
        title: "Ingress packet drops from TM perspective"
    },
    [pb.PROBE_TM_EGRESS_DROP_PKT]: {
        yAxisName: pb.Y_AXIS_NAME_PKT_COUNT,
        title: "Egress packet drops from TM perspective"
    },
    [pb.PROBE_TM_INRESS_USAGE_CELLS]: {
        yAxisName: pb.Y_AXIS_NAME_CELL_COUNT,
        title: "Port usage count in terms of number of memory cells usage from TM ingress perspective"
    },
    [pb.PROBE_TM_ERESS_USAGE_CELLS]: {
        yAxisName: pb.Y_AXIS_NAME_CELL_COUNT,
        title: "Port usage count in terms of number of memory cells usage from TM egress perspective"
    },
    [pb.PROBE_TM_PIPE_TOTAL_BUF_DROP]: {
        yAxisName: pb.Y_AXIS_NAME_PKT_COUNT,
        title: "Number of packets which were dropped because of buffer full condition"
    },
    [pb.PROBE_TM_PIPE_IG_FULL_BUF]: {
        yAxisName: pb.Y_AXIS_NAME_PKT_COUNT,
        title: "The number of packets which were dropped because of buffer full condition on ingress side"
    },
    [pb.PROBE_TM_PIPE_EG_DROP_PKT]: {
        yAxisName: pb.Y_AXIS_NAME_PKT_COUNT,
        title: "The total number of packets which were dropped on egress side"
    },
    [pb.PROBE_NEO_TPT_MTT_L0_MISS]: {
        yAxisName: pb.Y_AXIS_NAME_EVENTS_RATE,
        title: "Level 0 MTT Cache Miss"
    },
    [pb.PROBE_NEO_TPT_MTT_L1_MISS]: {
        yAxisName: pb.Y_AXIS_NAME_EVENTS_RATE,
        title: "Level 1 MTT Cache Miss"
    },
    [pb.PROBE_NEO_TPT_MPT_L0_MISS]: {
        yAxisName: pb.Y_AXIS_NAME_EVENTS_RATE,
        title: "Level 0 MPT Cache Miss"
    },
    [pb.PROBE_NEO_TPT_MPT_L1_MISS]: {
        yAxisName: pb.Y_AXIS_NAME_EVENTS_RATE,
        title: "Level 1 MPT Cache Miss"
    },
    [pb.PROBE_NEO_PCI_BP]: {
        yAxisName: pb.Y_AXIS_NAME_CYCLES_RATE,
        title: "PCIe Internal Back Pressure"
    },
    [pb.PROBE_NEO_ICM_MISS]: {
        yAxisName: pb.Y_AXIS_NAME_EVENTS_RATE,
        title: "ICM Cache Miss"
    },
    [pb.PROBE_NEO_RX_FULL_0]: {
        yAxisName: pb.Y_AXIS_NAME_CYCLES_RATE,
        title: "RX Packet Buffer Full Port 0"
    },
    [pb.PROBE_NEO_RX_FULL_1]: {
        yAxisName: pb.Y_AXIS_NAME_CYCLES_RATE,
        title: "RX Packet Buffer Full Port 1"
    },
    [pb.PROBE_NEO_WQE_MISS]: {
        yAxisName: pb.Y_AXIS_NAME_EVENTS_RATE,
        title: "Receive WQE Cache Miss"
    },
    [pb.PROBE_NEO_TX_BW]: {
        yAxisName: pb.Y_AXIS_NAME_GIGABYTE_RATE,
        title: "TX Bandwidth"
    },
    [pb.PROBE_NEO_RX_BW]: {
        yAxisName: pb.Y_AXIS_NAME_GIGABYTE_RATE,
        title: "RX Bandwidth",
        tickFormat: ".1s"
    },
    [pb.PROBE_NEO_TX_PKT]: {
        yAxisName: pb.Y_AXIS_NAME_PKT_RATE,
        title: "TX Packet Rate",
        tickFormat: ".1s"
    },
    [pb.PROBE_NEO_RX_PKT]: {
        yAxisName: pb.Y_AXIS_NAME_PKT_RATE,
        title: "RX Packet Rate"
    },
    [pb.PROBE_NEO_PCI_OUT_BW]: {
        yAxisName: pb.Y_AXIS_NAME_GIGABYTE_RATE,
        title: "PCIe Outbound Used BW"
    },
    [pb.PROBE_NEO_PCI_IN_BW]: {
        yAxisName: pb.Y_AXIS_NAME_GIGABYTE_RATE,
        title: "PCIe Inbound Used BW"
    },
    [pb.PROBE_ETHTOOL_RX_DISCARD]: {
        yAxisName: pb.Y_AXIS_NAME_PKT_COUNT,
        title: "RX packet discards"
    },
    [pb.PROBE_ETHTOOL_TX_DISCARD]: {
        yAxisName: pb.Y_AXIS_NAME_PKT_COUNT,
        title: "TX packet discards"
    },
    [pb.PROBE_ETHTOOL_RX_PAUSE]: {
        yAxisName: pb.Y_AXIS_NAME_PKT_COUNT,
        title: "Link layer pause frames received"
    },
    [pb.PROBE_ETHTOOL_TX_PAUSE]: {
        yAxisName: pb.Y_AXIS_NAME_PKT_COUNT,
        title: "Link layer pause frames sent"
    },
    [pb.PROBE_ETHTOOL_RX_OOB]: {
        yAxisName: pb.Y_AXIS_NAME_EVENTS_COUNT,
        title: "Out of buffer events for RX"
    },
}

export const getPifinaChartConfigByMetricName = (metricName: string): PIFINA_CHART_CONF_ITEM  => {
    if (metricName in PIFINA_PROBE_CHART_CFG) {
        return PIFINA_PROBE_CHART_CFG[metricName];
    }
    return {
        title: "Unknown chart",
        yAxisName: "Unknown scale",
        tickFormat: "s"
    }
}

export const getTickFormatFromPifinaChartConfig = (metricName: string): string => {
    const confItem = getPifinaChartConfigByMetricName(metricName);
    // return default value if tickFormat does not exists
    return confItem?.tickFormat ?? "s"
}
