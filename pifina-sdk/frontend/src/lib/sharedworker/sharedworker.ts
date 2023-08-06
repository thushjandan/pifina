import type { DTOTelemetryMessage } from "$lib/models/EndpointModel";

let evtSource: EventSource;
let ports: MessagePort[] = [];
const { MODE } = import.meta.env;
const evtSourceURL = MODE === 'development' ? 'https://localhost:8655' : ''

const evtSourceMessage = function(event: MessageEvent) {
    let dataobj: DTOTelemetryMessage = JSON.parse(event.data);
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