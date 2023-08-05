export interface EndpointModel {
    name: string
    type: EndpointType
    groupId: number
    address: string
    port: number
}

export enum EndpointType {
    HOSTTYPE_TOFINO = "HOSTTYPE_TOFINO",
    HOSTTYPE_NIC = "HOSTTYPE_NIC"
}