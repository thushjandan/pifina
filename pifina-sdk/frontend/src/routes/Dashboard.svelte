<script lang="ts">
	import Chart from './Chart.svelte';
	import * as Plot from "@observablehq/plot";
	import type { DTOPifinaMetricItem, MetricData, MetricItem } from '../lib/models/MetricItem';
	import { PifinaMetricName } from '$lib/models/metricTypes';
    export let endpoints: string[];

	let cliendScreenWidth;
    let selectedEndpoint: string = endpoints[0];
	let metricData: MetricData = {};

	const evtSourceMessage = function(event: MessageEvent) {
		let dataobj: DTOPifinaMetricItem[] = JSON.parse(event.data);
		dataobj.forEach(item => {
			metricData[item.metricName] = metricData[item.metricName] || [];
			metricData[item.metricName].push({timestamp: new Date(item.timestamp), value: item.value, sessionId: item.sessionId});
		})
		if (metricData[PifinaMetricName.INGRESS_MATCH_CNT].length > 500) {
			let deleteCount = metricData[PifinaMetricName.INGRESS_MATCH_CNT].length - 500;
			metricData[PifinaMetricName.INGRESS_MATCH_CNT].splice(0, deleteCount);
		}
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
	let plotOptions: Plot.PlotOptions = { x: { domain: [100, 0] }, grid: true };
</script>

<div class="grid grid-cols-4">
	<div class="sm:col-span-1">
		<label for="username" class="block text-sm font-medium leading-6 text-gray-900">Choose a monitoring target:</label>
		<div class="mt-2">
			<select bind:value={selectedEndpoint} on:change={onEndpointChange} placeholder="Placeholder" class="px-3 py-3 placeholder-slate-300 text-slate-600 relative bg-white bg-white rounded text-sm border-0 shadow outline-none focus:outline-none focus:ring w-full">
				{#each endpoints as endpoint }
				<option value={endpoint}>{endpoint}</option>
				{/each}
			</select>
		</div>
	</div>
</div>
{#if PifinaMetricName.INGRESS_MATCH_CNT in metricData }
<div bind:clientWidth={cliendScreenWidth} class="mt-8 w-full">
	<h2>First ingress byte counter</h2>
	<Chart options={{
		x: {
			label: "Timestamp",
			tickFormat: (value, _) => (value.toLocaleString(undefined, {
				hour: 'numeric',
				minute: 'numeric',
				second: 'numeric'
			})),
		},
		y: {
			label: "(bytes/sec)",
			grid: true
		},
		width: cliendScreenWidth,
		color: {legend: true, type: "categorical"},
		marks: [
			Plot.line(metricData[PifinaMetricName.INGRESS_MATCH_CNT], {x: "timestamp", y: "value", stroke: "sessionId", marker: "dot"}),
			Plot.tickY(metricData[PifinaMetricName.INGRESS_MATCH_CNT], {y: "value", title: (d) => (`${d.value} bytes/sec`), strokeWidth: 12, opacity: 0.01, stroke: "white"})
		]
	}} />
</div>
{/if}
<div class="mt-8">
	<h2>Ingress & Egress Packet Counter</h2>
	<Chart options={plotOptions} />
</div>
<div class="mt-8">
	<h2>Ingress & Egress Packet Header Size</h2>
	<Chart options={plotOptions} />
</div>