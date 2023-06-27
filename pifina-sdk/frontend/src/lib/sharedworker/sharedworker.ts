import type { DTOPifinaMetricItem } from "$lib/models/MetricItem";

let evtSource: EventSource;
let ports: MessagePort[] = [];

const evtSourceMessage = function(event: MessageEvent) {
    let dataobj: DTOPifinaMetricItem[] = JSON.parse(event.data);
    ports.forEach(port => {
        port.postMessage(dataobj);
    });
}

const createEventSource = (endpoint: string) => {
    if (typeof evtSource === "undefined") {
        evtSource = new EventSource(`https://localhost:8655/api/v1/events?stream=${endpoint}`);
        evtSource.onmessage = evtSourceMessage;
        return
    }

    if (evtSource.url !== `https://localhost:8655/api/v1/events?stream=${endpoint}`) {
        evtSource.close()
        evtSource = new EventSource(`https://localhost:8655/api/v1/events?stream=${endpoint}`);
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
                createEventSource(workerData.endpoint);
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