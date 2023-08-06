import { PIFINA_DEFAULT_PROBE_CHART_ORDER, PIFINA_TM_CHART_ORDER } from "./metricNames";

export const PIFINA_DASHBOARD_CONF: PIFINA_DASHBOARD_CONF_TYPE = {
    HOSTTYPE_TOFINO: [
        {
            key: "MAIN_CHARTS",
            title: "Default Probes",
            type: "static",
            charts: PIFINA_DEFAULT_PROBE_CHART_ORDER,
            disableSessionFilter: false
        },
        {
            key: "APP_REG_CHARTS",
            title: "Application owned registers",
            type: "list",
            groupName: "appRegister",
            disableSessionFilter: true
        },
        {
            key: "EXTRA_PROBES_CHARTS",
            title: "Extra probes",
            type: "list",
            groupName: "extraProbes",
            disableSessionFilter: false
        },
        {
            key: "TM_CHARTS",
            title: "Traffic Manager",
            type: "static",
            charts: PIFINA_TM_CHART_ORDER,
            disableSessionFilter: true
        }        
    ],
    HOSTTYPE_NIC: []
}

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