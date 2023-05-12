<script lang="ts">
	import Chart from './Chart.svelte';
	import * as Plot from "@observablehq/plot";
	import type { DTOPifinaMetricItem, MetricData, MetricItem } from '../lib/models/MetricItem';
	import { PROBE_EGRESS_END_CNT_BYTE, PROBE_EGRESS_START_CNT_BYTE, PROBE_EGRESS_START_CNT_PKTS, PROBE_INGRESS_END_HDR_BYTE, PROBE_INGRESS_MATCH_CNT_BYTE, PROBE_INGRESS_MATCH_CNT_PKT, PROBE_INGRESS_START_HDR_BYTE } from '$lib/models/metricNames';
    export let endpoints: string[];

	let cliendScreenWidth;
    let selectedEndpoint: string = endpoints[0];
	let sessionIds = new Set<number>();
	let selectedSessionIds: number[] = [];
	let metricData: MetricData = {};

	const evtSourceMessage = function(event: MessageEvent) {
		let dataobj: DTOPifinaMetricItem[] = JSON.parse(event.data);
		dataobj.forEach(item => {
			const key = `${item.metricName}${item.type}`;
			// check if key exists. If not, create a new list.
			if (!(key in metricData)) {
				metricData[key] = [];
			}
			if (!sessionIds.has(item.sessionId)) {
				sessionIds.add(item.sessionId);
			}
			metricData[key].push({timestamp: new Date(item.timestamp), value: item.value, sessionId: item.sessionId, type: item.type});
		})
		// Default initialization of filters
		if (selectedSessionIds.length == 0) {
			for (const sessionId of sessionIds.values()) {
				selectedSessionIds.push(sessionId);
			}
		}

		// Limit series length
		const metricSizeLimit = (selectedSessionIds.length + 1) * 500;
		for (const mapkey in metricData) {
			if (metricData[mapkey].length > metricSizeLimit) {
				metricData[mapkey].splice(0, metricData[mapkey].length - metricSizeLimit);
			}
		}
		//console.log(dataobj);
		// Force rerender
		metricData = metricData;
	}

	let evtSource = new EventSource(`https://localhost:8655/api/v1/events?stream=${endpoints[0]}`);
	evtSource.onmessage = evtSourceMessage;


	const onEndpointChange = () => {
		if (typeof evtSource !== "undefined") {
			// Close previous event source
			evtSource.close()
		}
		evtSource = new EventSource(`https://localhost:8655/api/v1/events?stream=${endpoints[0]}`);
		evtSource.onmessage = evtSourceMessage;
	}

	const xScaleOptions: Plot.ScaleOptions = {
		label: "Timestamp",
		tickFormat: (value, _) => (value.toLocaleString(undefined, {
			hour: 'numeric',
			minute: 'numeric',
			second: 'numeric'
		})),
	};
</script>

<div class="grid grid-cols-4">
	<div class="sm:col-span-1">
		<label for="target" class="block text-sm font-medium leading-6 text-gray-900">Choose a monitoring target:</label>
		<div class="mt-2">
			<select bind:value={selectedEndpoint} on:change={onEndpointChange} name="target" class="px-3 py-3 placeholder-slate-300 text-slate-600 relative bg-white bg-white rounded text-sm border-0 shadow outline-none focus:outline-none focus:ring w-full">
				{#each endpoints as endpoint }
				<option value={endpoint}>{endpoint}</option>
				{/each}
			</select>
		</div>
	</div>
</div>
<div class="divide-y divide-solid">
	{#if PROBE_INGRESS_MATCH_CNT_BYTE in metricData }
	<div class="mt-8">
		<div class="sm:col-span-1">
			<label for="sessionIds" class="block text-sm font-medium leading-6 text-gray-900">Filter by session ID:</label>
			<div class="mt-2 flex flex-row">
				{#each [...sessionIds.values()] as sessionId}
						<div class="items-center">
							<input type=checkbox bind:group={selectedSessionIds} name="sessionIds" value={sessionId} class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600" />
							<label for="comments" class="ml-1 mr-4 font-medium text-gray-900">{sessionId}</label>
						</div>
				{/each}
			</div>
		</div>
	</div>
	<div bind:clientWidth={cliendScreenWidth} class="mt-8 pt-4 w-full">
		<h2>Start ingress byte counter</h2>
		<Chart options={{
			x: xScaleOptions,
			y: {
				label: "(bytes/sec)",
				grid: true
			},
			width: cliendScreenWidth,
			color: {legend: true, type: "categorical"},
			marks: [
				Plot.line(metricData[PROBE_INGRESS_MATCH_CNT_BYTE], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), x: "timestamp", y: "value", stroke: "sessionId", marker: "dot"}),
				Plot.tickY(metricData[PROBE_INGRESS_MATCH_CNT_BYTE], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), y: "value", title: (d) => (`${d.value} bytes/sec`), strokeWidth: 12, opacity: 0.001, stroke: "white"})
			]
		}} />
	</div>
	{/if}
	
	{#if PROBE_EGRESS_START_CNT_BYTE in metricData && PROBE_EGRESS_END_CNT_BYTE in metricData }
	<div bind:clientWidth={cliendScreenWidth} class="mt-8 pt-4">
		<h2>Start & end egress byte counter</h2>
		<Chart options={{
			x: xScaleOptions,
			y: {
				label: "(bytes/sec)",
				grid: true
			},
			width: cliendScreenWidth,
			color: {legend: true, type: "categorical"},
			marks: [
				Plot.line(metricData[PROBE_EGRESS_START_CNT_BYTE], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), x: "timestamp", y: "value", stroke: (d) => `Start: ${d.sessionId}`, marker: "dot"}),
				Plot.tickY(metricData[PROBE_EGRESS_START_CNT_BYTE], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), y: "value", title: (d) => (`${d.value} bytes/sec`), strokeWidth: 12, opacity: 0.001, stroke: (d) => `Start: ${d.sessionId}`}),
				Plot.line(metricData[PROBE_EGRESS_END_CNT_BYTE], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), x: "timestamp", y: "value", stroke: (d) => `End: ${d.sessionId}`, marker: "dot"}),
				Plot.tickY(metricData[PROBE_EGRESS_END_CNT_BYTE], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), y: "value", title: (d) => (`${d.value} bytes/sec`), strokeWidth: 12, opacity: 0.001, stroke: (d) => `End: ${d.sessionId}`})
			]
		}} />
	</div>
	{/if}
	
	{#if PROBE_EGRESS_START_CNT_PKTS in metricData && PROBE_INGRESS_MATCH_CNT_PKT in metricData }
	<div bind:clientWidth={cliendScreenWidth} class="mt-8 pt-4">
		<h2>Ingress & egress packet counter</h2>
		<Chart options={{
			x: xScaleOptions,
			y: {
				label: "(pkts/sec)",
				grid: true
			},
			width: cliendScreenWidth,
			color: {legend: true, type: "categorical"},
			marks: [
				Plot.line(metricData[PROBE_INGRESS_MATCH_CNT_PKT], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), x: "timestamp", y: "value", stroke: (d) => `Ingress: ${d.sessionId}`, marker: "dot"}),
				Plot.line(metricData[PROBE_EGRESS_START_CNT_PKTS], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), x: "timestamp", y: "value", stroke: (d) => `Egress: ${d.sessionId}`, marker: "dot"}),
				Plot.tickY(metricData[PROBE_INGRESS_MATCH_CNT_PKT], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), y: "value", title: (d) => (`${d.value} pkts/sec`), strokeWidth: 12, opacity: 0.001, stroke: (d) => `Ingress: ${d.sessionId}`}),
				Plot.tickY(metricData[PROBE_EGRESS_START_CNT_PKTS], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), y: "value", title: (d) => (`${d.value} pkts/sec`), strokeWidth: 12, opacity: 0.001, stroke: (d) => `Egress: ${d.sessionId}`})
			]
		}} />
	</div>
	{/if}
	
	{#if PROBE_INGRESS_START_HDR_BYTE in metricData && PROBE_INGRESS_END_HDR_BYTE in metricData }
	<div bind:clientWidth={cliendScreenWidth} class="mt-8 pt-4">
		<h2>Ingress packet header size</h2>
		<Chart options={{
			x: xScaleOptions,
			y: {
				label: "(bytes/sec)",
				grid: true
			},
			width: cliendScreenWidth,
			color: {legend: true, type: "categorical"},
			marks: [
				Plot.line(metricData[PROBE_INGRESS_START_HDR_BYTE], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), x: "timestamp", y: "value", stroke: (d) => `Start: ${d.sessionId}`, marker: "dot"}),
				Plot.line(metricData[PROBE_INGRESS_END_HDR_BYTE], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), x: "timestamp", y: "value", stroke: (d) => `End: ${d.sessionId}`, marker: "dot"}),
				Plot.tickY(metricData[PROBE_INGRESS_START_HDR_BYTE], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), y: "value", title: (d) => (`${d.value} bytes/sec`), strokeWidth: 12, opacity: 0.001, stroke: (d) => `Start: ${d.sessionId}`}),
				Plot.tickY(metricData[PROBE_INGRESS_END_HDR_BYTE], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), y: "value", title: (d) => (`${d.value} bytes/sec`), strokeWidth: 12, opacity: 0.001, stroke: (d) => `End: ${d.sessionId}`})
			]
		}} />
	</div>
	{/if}
</div>