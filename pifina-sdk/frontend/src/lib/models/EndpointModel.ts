import type { DTOPifinaMetricItem } from "./MetricItem"

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