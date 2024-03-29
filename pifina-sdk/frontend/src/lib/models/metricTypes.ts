// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

export enum PifinaMetricName {
    INGRESS_MATCH_CNT = "PF_INGRESS_MATCH_CNT",
    INGRESS_START_HDR = "PF_INGRESS_START_HDR_SIZE",
    INGRESS_END_HDR = "PF_INGRESS_END_HDR_SIZE",
    EGRESS_START_CNT = "PF_EGRESS_START_CNT",
    EGRESS_END_CNT = "PF_EGRESS_END_CNT",
    INGRESS_JITTER_AVG = "PF_INGRESS_JITTER_AVG"
}

export enum MetricTypes {
    BYTES = "METRIC_BYTES",
    PKTS = "METRIC_PKTS",
    EXT_VALUE = "METRIC_EXT_VALUE"
}