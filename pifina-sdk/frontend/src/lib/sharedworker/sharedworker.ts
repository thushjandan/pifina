import type { DTOTelemetryMessage } from "$lib/models/EndpointModel";
import { PifinaMetricName } from "$lib/models/metricTypes";

let evtSource: EventSource;
let ports: MessagePort[] = [];
const { MODE } = import.meta.env;
const evtSourceURL = MODE === 'development' ? 'https://localhost:8655' : ''

const evtSourceMessage = function(event: MessageEvent) {
    let dataobj: DTOTelemetryMessage = JSON.parse(event.data);
    dataobj.metrics = dataobj.metrics.map(item => {
        // Convert nano seconds to miliseconds
        if (item.metricName == PifinaMetricName.INGRESS_JITTER_AVG) {
            if (item.value > 0) {
                item.value = Math.round(item.value / 1000);
            }
        }
        return item
    })
    ports.forEach(port => {
        port.postMessage(dataobj);
    });
}

const createEventSource = (groupId: string) => {
    const evtSourceUrl = `${evtSourceURL}/api/v1/events?stream=group${groupId}`;
    if (typeof evtSource === "undefined") {
        evtSource = new EventSource(evtSourceUrl);
        evtSource.onmessage = evtSourceMessage;
        return
    }

    if (evtSource.url !== evtSourceUrl || evtSource.readyState === EventSource.CLOSED) {
        evtSource.close()
        evtSource = new EventSource(evtSourceUrl);
        evtSource.onmessage = evtSourceMessage;
    }
}

onconnect = (e: MessageEvent) => {
    const port = e.ports[0];
    ports.push(port);
    port.onmessage = (e: MessageEvent) => {
        const workerData = e.data;
        switch (workerData.status) {
            case "CONNECT":
                createEventSource(workerData.groupId);
                break;
            case "CLOSE":
                if (typeof evtSource !== "undefined") {
                    evtSource.close()
                }
                break;
            default:
                break;
        }
    };

    port.start();
}