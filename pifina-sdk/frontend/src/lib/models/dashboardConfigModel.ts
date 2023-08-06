export interface PIFINA_DASHBOARD_CONF_ITEM {
    key: string
    title: string
    type: string
    charts?: (string[]| string)[]
    groupName?: string
    disableSessionFilter: boolean
}

export interface PIFINA_DASHBOARD_CONF_TYPE {
    HOSTTYPE_TOFINO: PIFINA_DASHBOARD_CONF_ITEM[]
    HOSTTYPE_NIC: PIFINA_DASHBOARD_CONF_ITEM[]
}