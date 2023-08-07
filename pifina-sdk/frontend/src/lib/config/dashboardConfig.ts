// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import type { PIFINA_DASHBOARD_CONF_TYPE } from "$lib/models/dashboardConfigModel";
import { PIFINA_DEFAULT_PROBE_CHART_ORDER, PIFINA_ETHTOOL_CHART_ORDER, PIFINA_NEO_CHART_ORDER, PIFINA_TM_CHART_ORDER } from "./chartOrderConfig";

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
    HOSTTYPE_NIC: [
        {
            key: "ETHTOOL_CHARTS",
            title: "Ethtool",
            type: "static",
            charts: PIFINA_ETHTOOL_CHART_ORDER,
            disableSessionFilter: true
        },
        {
            key: "NEOHOST_CHARTS",
            title: "NEO-Host",
            type: "static",
            charts: PIFINA_NEO_CHART_ORDER,
            disableSessionFilter: true
        }
    ]
}
