// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import type { DTOPifinaMetricItem } from "./metricItem"

export interface EndpointModel {
    name: string
    type: EndpointType
    groupId: number
    address: string
    port: number
}

export interface DTOTelemetryMessage {
    source: string
    type: EndpointType
    groupId: number
    metrics: DTOPifinaMetricItem[]
}

export enum EndpointType {
    HOSTTYPE_TOFINO = "HOSTTYPE_TOFINO",
    HOSTTYPE_NIC = "HOSTTYPE_NIC"
}