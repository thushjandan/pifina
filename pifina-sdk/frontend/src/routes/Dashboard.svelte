<script lang="ts">
	import Chart from './Chart.svelte';
	import * as Plot from "@observablehq/plot";
	import type { DTOPifinaMetricItem, MetricData } from '../lib/models/MetricItem';
	import { PROBE_EGRESS_END_CNT_BYTE, PROBE_EGRESS_START_CNT_BYTE, PROBE_EGRESS_START_CNT_PKTS, PROBE_INGRESS_END_HDR_BYTE, PROBE_INGRESS_JITTER, PROBE_INGRESS_MATCH_CNT_BYTE, PROBE_INGRESS_MATCH_CNT_PKT, PROBE_INGRESS_START_HDR_BYTE, PROBE_TM_EGRESS_DROP_PKT, PROBE_TM_ERESS_USAGE_CELLS, PROBE_TM_INGRESS_DROP_PKT, PROBE_TM_INRESS_USAGE_CELLS, PROBE_TM_PIPE_EG_DROP_PKT, PROBE_TM_PIPE_IG_FULL_BUF, PROBE_TM_PIPE_TOTAL_BUF_DROP } from '$lib/models/metricNames';
	import { onDestroy } from 'svelte';
	import { ChartMenuCategoryModel } from '$lib/models/ChartMenuCategory';
    export let endpoints: string[];

	let cliendScreenWidth;
    let selectedEndpoint: string = endpoints[0];
	let sessionIds = new Set<number>();
	let selectedSessionIds: number[] = [];
	let sessionIdFilterIsDirty = false;
	let metricData: MetricData = {};
	let appRegister = new Set<string>();
	let extraProbeNames = new Set<string>();
	let tmMetrics = new Set<string>();
	let selectedChartCategory = ChartMenuCategoryModel.MAIN_CHARTS;
	let isEnabled = true;

	const evtSourceMessage = function(event: MessageEvent) {
		let dataobj: DTOPifinaMetricItem[] = JSON.parse(event.data);
		dataobj.forEach(item => {
			let key = `${item.metricName}${item.type}`;
			// Check if it's a metric from a default probe
			if (!item.metricName.startsWith("PF_")) {
				appRegister.add(item.metricName);
				key = item.metricName;
			}
			if (item.metricName.startsWith("PF_TM_")) {
				tmMetrics.add(item.metricName)
				key = item.metricName;
			}
			if (item.metricName.startsWith("PF_EXTRA")) {
				extraProbeNames.add(item.metricName);
				key = item.metricName;
			}
			// check if key exists. If not, create a new list.
			if (!(key in metricData)) {
				metricData[key] = [];
			}
			if (!sessionIds.has(item.sessionId) && !item.metricName.startsWith("PF_TM_")) {
				sessionIds.add(item.sessionId);
			}
			metricData[key].push({timestamp: new Date(item.timestamp), value: item.value, sessionId: item.sessionId, type: item.type});
		})

		// Default initialization of filters
		if (!sessionIdFilterIsDirty && selectedSessionIds.length == 0) {
			for (const sessionId of sessionIds.values()) {
				selectedSessionIds.push(sessionId);
			}
		}

		// Limit series length
		for (const mapkey in metricData) {
			let mapKeySessioId = new Set<number>();
			const groupBySessionId = metricData[mapkey].forEach(item => {
				mapKeySessioId.add(item.sessionId);
			});
			const metricSizeLimit = (mapKeySessioId.size + 1) * 20;
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

	const startEventStreaming = () => {
		evtSource = new EventSource(`https://localhost:8655/api/v1/events?stream=${endpoints[0]}`);
		evtSource.onmessage = evtSourceMessage;		
	}

	const toggleEventStreaming = () => {
		if (isEnabled && typeof evtSource !== "undefined") {
			// Close previous event source
			evtSource.close()
		}
		if (!isEnabled) {
			startEventStreaming();
		}
		isEnabled = !isEnabled;
	}

	const isMainChartSelected = () => selectedChartCategory == ChartMenuCategoryModel.MAIN_CHARTS;
	const isTMChartSelected = () => selectedChartCategory == ChartMenuCategoryModel.TM_CHARTS;

	const xScaleOptions: Plot.ScaleOptions = {
		label: "Timestamp",
		tickSpacing: 150,
		tickFormat: (value, _) => (value.toLocaleString(undefined, {
			hour: 'numeric',
			minute: 'numeric',
			second: 'numeric'
		})),
	};
	onDestroy(() => {
		if (typeof evtSource !== 'undefined') {
			evtSource.close();
		}
	});
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
	<div class="sm:col-span-3 justify-self-end">
		<div class="mt-2 mr-4">
			<button on:click={toggleEventStreaming} class:bg-orange-600={isEnabled} class:hover:bg-orange-800={isEnabled} class:bg-green-600={!isEnabled} class:hover:bg-green-600={!isEnabled} class="text-white text-center font-medium rounded-lg text-sm w-full sm:w-auto px-3 py-2.5 mr-2">
				{#if isEnabled}
				<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="inline w-4 h-4 mr-2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25v13.5m-7.5-13.5v13.5" />
				</svg>				  
				Pause
				{:else}
				<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="inline w-4 h-4 mr-2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.348a1.125 1.125 0 010 1.971l-11.54 6.347a1.125 1.125 0 01-1.667-.985V5.653z" />
				</svg>
				Start
				{/if}
			</button>
		</div>
	</div>
</div>
{#key selectedChartCategory }
<div class="mt-6 text-sm font-medium text-center text-gray-500 border-b border-gray-200 dark:text-gray-400 dark:border-gray-700">
    <ul class="flex flex-wrap -mb-px">
        <li class="mr-2">
            <a href={null} on:click={() => selectedChartCategory = ChartMenuCategoryModel.MAIN_CHARTS} class="inline-block p-4 border-b-2 rounded-t-lg hover:text-gray-600 hover:border-gray-300" class:border-transparent={!isMainChartSelected()} class:text-indigo-600={isMainChartSelected()} 
				class:border-blue-600={isMainChartSelected()} class:hover:text-gray-600={!isMainChartSelected()} class:hover:border-gray-300={!isMainChartSelected()}>
				Default Probes
			</a>
        </li>
        <li class="mr-2">
            <a href={null} on:click={() => selectedChartCategory = ChartMenuCategoryModel.TM_CHARTS} class="inline-block p-4 border-b-2 rounded-t-lg hover:text-gray-600 hover:border-gray-300" class:border-transparent={!isTMChartSelected()} class:text-indigo-600={isTMChartSelected()} 
				class:border-blue-600={isTMChartSelected()} class:hover:text-gray-600={!isTMChartSelected()} class:hover:border-gray-300={!isTMChartSelected()}>
				Traffic Manager</a>
        </li>
    </ul>
</div>
{#if isMainChartSelected() == true }
<div class="divide-y divide-solid">
	{#if PROBE_INGRESS_MATCH_CNT_BYTE in metricData }
	<div class="mt-8">
		<div class="sm:col-span-1">
			<label for="sessionIds" class="block text-sm font-medium leading-6 text-gray-900">Filter by session ID:</label>
			<div class="mt-2 flex flex-row">
				{#each [...sessionIds.values()] as sessionId}
						<div class="items-center">
							<input type=checkbox bind:group={selectedSessionIds} on:change={() => sessionIdFilterIsDirty = true} name="sessionIds" value={sessionId} class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600" />
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
				Plot.tip(metricData[PROBE_INGRESS_MATCH_CNT_BYTE], Plot.pointerX({x: "timestamp", y: "value", channels: {sessionId: "sessionId"}, filter: (d) => (selectedSessionIds.includes(d.sessionId))})),
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
				Plot.line(metricData[PROBE_EGRESS_END_CNT_BYTE], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), x: "timestamp", y: "value", stroke: (d) => `End: ${d.sessionId}`, marker: "dot"}),
				Plot.tip(metricData[PROBE_EGRESS_START_CNT_BYTE], Plot.pointerX({x: "timestamp", y: "value", channels: {sessionId: "sessionId"}, title: (d) => `Start byte counter for session ${d.sessionId}\n\n${d.value} bytes/sec`, filter: (d) => (selectedSessionIds.includes(d.sessionId))})),
				Plot.tip(metricData[PROBE_EGRESS_END_CNT_BYTE], Plot.pointerX({x: "timestamp", y: "value", channels: {sessionId: "sessionId"}, title: (d) => `End byte counter for session ${d.sessionId}\n\n${d.value} bytes/sec`, filter: (d) => (selectedSessionIds.includes(d.sessionId))})),
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
	{#if PROBE_INGRESS_JITTER in metricData }
	<div bind:clientWidth={cliendScreenWidth} class="mt-8 pt-4">
		<h2>Moving average ingress jitter</h2>
		<Chart options={{
			x: xScaleOptions,
			y: {
				label: "ms",
				grid: true
			},
			width: cliendScreenWidth,
			color: {legend: true, type: "categorical"},
			marks: [
				Plot.line(metricData[PROBE_INGRESS_JITTER], {filter: (d) => (selectedSessionIds.includes(d.sessionId)), x: "timestamp", y: "value", stroke: (d) => `Start: ${d.sessionId}`, marker: "dot"}),
				Plot.tip(metricData[PROBE_INGRESS_JITTER], Plot.pointerX({x: "timestamp", y: "value", channels: {sessionId: "sessionId"}, filter: (d) => (selectedSessionIds.includes(d.sessionId))})),
			]
		}} />
	</div>
	{/if}
	{#each [...extraProbeNames.values()] as entry }
	<div bind:clientWidth={cliendScreenWidth} class="mt-8 pt-4">
		<h2>{entry}</h2>
		<Chart options={{
			x: xScaleOptions,
			y: {
				label: "bytes/sec",
				grid: true
			},
			width: cliendScreenWidth,
			color: {legend: true, type: "categorical"},
			marks: [
				Plot.line(metricData[entry], {x: "timestamp", y: "value", stroke: (d) => `${d.sessionId}`, marker: "dot"}),
				Plot.tip(metricData[entry], Plot.pointerX({x: "timestamp", y: "value", channels: {sessionId: "sessionId"}, filter: (d) => (selectedSessionIds.includes(d.sessionId))})),
			]
		}} />
	</div>
	{/each}
	{#each [...appRegister.values()] as entry }
	<div bind:clientWidth={cliendScreenWidth} class="mt-8 pt-4">
		<h2>{entry} register (app-owned)</h2>
		<Chart options={{
			x: xScaleOptions,
			y: {
				label: "current",
				grid: true
			},
			width: cliendScreenWidth,
			color: {legend: true, type: "categorical"},
			marks: [
				Plot.line(metricData[entry], {x: "timestamp", y: "value", stroke: (d) => `${d.sessionId}`, marker: "dot"}),
				Plot.tickY(metricData[entry], {y: "value", title: (d) => (`current: ${d.value}`), strokeWidth: 12, opacity: 0.001, stroke: (d) => `${d.sessionId}`}),
			]
		}} />
	</div>		
	{/each}
</div>
{/if}
{#if isTMChartSelected() == true }
<div class="divide-y divide-solid">
	<div class="mt-8 grid md:grid-cols-2">
		{#if PROBE_TM_INGRESS_DROP_PKT in metricData }
		<div bind:clientWidth={cliendScreenWidth} class="mt-8 pt-4">
			<h2>Ingress & egress packet drops from TM perspective</h2>
			<Chart options={{
				x: xScaleOptions,
				y: {
					label: "pkts",
					grid: true
				},
				width: cliendScreenWidth,
				color: {legend: true, type: "categorical"},
				marks: [
					Plot.line(metricData[PROBE_TM_INGRESS_DROP_PKT], {x: "timestamp", y: "value", stroke: (d) => `ingress: ${d.sessionId}`, marker: "dot"}),
					Plot.line(metricData[PROBE_TM_EGRESS_DROP_PKT], {x: "timestamp", y: "value", stroke: (d) => `egress: ${d.sessionId}`, marker: "dot"}),
					Plot.tickY(metricData[PROBE_TM_INGRESS_DROP_PKT], {y: "value", title: (d) => (`${d.value} pkts`), strokeWidth: 12, opacity: 0.001, stroke: (d) => `ingress: ${d.sessionId}`}),
					Plot.tickY(metricData[PROBE_TM_EGRESS_DROP_PKT], {y: "value", title: (d) => (`${d.value} pkts`), strokeWidth: 12, opacity: 0.001, stroke: (d) => `egress: ${d.sessionId}`})
				]
			}} />
		</div>
		<div bind:clientWidth={cliendScreenWidth} class="mt-8 pt-4">
			<h2>Port usage count in terms of number of memory cells usage from TM ingress & egress perspective</h2>
			<Chart options={{
				x: xScaleOptions,
				y: {
					label: "pkts",
					grid: true
				},
				width: cliendScreenWidth,
				color: {legend: true, type: "categorical"},
				marks: [
					Plot.line(metricData[PROBE_TM_INRESS_USAGE_CELLS], {x: "timestamp", y: "value", stroke: (d) => `ingress: ${d.sessionId}`, marker: "dot"}),
					Plot.line(metricData[PROBE_TM_ERESS_USAGE_CELLS], {x: "timestamp", y: "value", stroke: (d) => `egress: ${d.sessionId}`, marker: "dot"}),
					Plot.tickY(metricData[PROBE_TM_INRESS_USAGE_CELLS], {y: "value", title: (d) => (`${d.value} pkts`), strokeWidth: 12, opacity: 0.001, stroke: (d) => `ingress: ${d.sessionId}`}),
					Plot.tickY(metricData[PROBE_TM_ERESS_USAGE_CELLS], {y: "value", title: (d) => (`${d.value} pkts`), strokeWidth: 12, opacity: 0.001, stroke: (d) => `egress: ${d.sessionId}`})
				]
			}} />
		</div>
		{/if}
		{#if PROBE_TM_PIPE_TOTAL_BUF_DROP in metricData }
		<div bind:clientWidth={cliendScreenWidth} class="mt-8 pt-4">
			<h2>Number of packets which were dropped because of buffer full condition</h2>
			<Chart options={{
				x: xScaleOptions,
				y: {
					label: "pkts",
					grid: true
				},
				width: cliendScreenWidth,
				color: {legend: true, type: "categorical"},
				marks: [
					Plot.line(metricData[PROBE_TM_PIPE_TOTAL_BUF_DROP], {x: "timestamp", y: "value", stroke: (d) => `pipeline: ${d.sessionId}`,  marker: "dot"}),
					Plot.tip(metricData[PROBE_TM_PIPE_TOTAL_BUF_DROP], Plot.pointerX({x: "timestamp", y: "value"})),
				]
			}} />
		</div>
		<div bind:clientWidth={cliendScreenWidth} class="mt-8 pt-4">
			<h2>The number of packets which were dropped because of buffer full condition on ingress side</h2>
			<Chart options={{
				x: xScaleOptions,
				y: {
					label: "pkts",
					grid: true
				},
				width: cliendScreenWidth,
				color: {legend: true, type: "categorical"},
				marks: [
					Plot.line(metricData[PROBE_TM_PIPE_IG_FULL_BUF], {x: "timestamp", y: "value", stroke: (d) => `pipeline: ${d.sessionId}`, marker: "dot"}),
					Plot.tip(metricData[PROBE_TM_PIPE_IG_FULL_BUF], Plot.pointerX({x: "timestamp", y: "value"})),
				]
			}} />
		</div>
		<div bind:clientWidth={cliendScreenWidth} class="mt-8 pt-4">
			<h2>The total number of packets which were dropped on egress side.</h2>
			<Chart options={{
				x: xScaleOptions,
				y: {
					label: "pkts",
					grid: true
				},
				width: cliendScreenWidth,
				color: {legend: true, type: "categorical"},
				marks: [
					Plot.line(metricData[PROBE_TM_PIPE_EG_DROP_PKT], {x: "timestamp", y: "value", stroke: (d) => `pipeline: ${d.sessionId}`, marker: "dot"}),
					Plot.tip(metricData[PROBE_TM_PIPE_EG_DROP_PKT], Plot.pointerX({x: "timestamp", y: "value"}))
				]
			}} />
		</div>
		{/if}
	</div>
</div>
{/if}
{/key}